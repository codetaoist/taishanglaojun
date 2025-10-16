import React, { useState, useEffect } from 'react';
import { Card, Tree, Input, Tag, Space, Typography, Divider, Row, Col, Button, Tooltip, message, Spin } from 'antd';
import { SearchOutlined, ApiOutlined, FolderOutlined, FileTextOutlined, LinkOutlined, ReloadOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import type { DataNode } from 'antd/es/tree';
import apiDocumentationService, { APICategory, APIEndpoint } from '../../services/apiDocumentationService';

const { Title, Text, Paragraph } = Typography;
const { Search } = Input;

const APICatalog: React.FC = () => {
  const { t } = useTranslation();
  const [searchValue, setSearchValue] = useState('');
  const [expandedKeys, setExpandedKeys] = useState<React.Key[]>([]);
  const [selectedKeys, setSelectedKeys] = useState<React.Key[]>([]);
  const [selectedEndpoint, setSelectedEndpoint] = useState<APIEndpoint | null>(null);
  const [loading, setLoading] = useState(false);
  const [categories, setCategories] = useState<APICategory[]>([]);
  const [endpoints, setEndpoints] = useState<APIEndpoint[]>([]);
  const [filteredEndpoints, setFilteredEndpoints] = useState<APIEndpoint[]>([]);

  // 加载数据
  const loadData = async () => {
    setLoading(true);
    try {
      const [categoriesRes, endpointsRes] = await Promise.all([
        apiDocumentationService.getCategories(1, 100),
        apiDocumentationService.getEndpoints(1, 1000)
      ]);
      
      setCategories(categoriesRes.categories);
      setEndpoints(endpointsRes.endpoints);
      setFilteredEndpoints(endpointsRes.endpoints);
      
      // 默认展开所有分类
      setExpandedKeys(categoriesRes.categories.map(cat => `category-${cat.id}`));
    } catch (error) {
      console.error('Failed to load API documentation:', error);
      message.error('加载API文档失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  // 搜索功能
  useEffect(() => {
    if (!searchValue.trim()) {
      setFilteredEndpoints(endpoints);
      return;
    }

    const filtered = endpoints.filter(endpoint => 
      endpoint.name.toLowerCase().includes(searchValue.toLowerCase()) ||
      endpoint.path.toLowerCase().includes(searchValue.toLowerCase()) ||
      endpoint.description.toLowerCase().includes(searchValue.toLowerCase())
    );
    setFilteredEndpoints(filtered);
  }, [searchValue, endpoints]);

  // 构建树形数据
  const buildTreeData = (): DataNode[] => {
    return categories.map(category => {
      const categoryEndpoints = filteredEndpoints.filter(endpoint => endpoint.category_id === category.id);
      
      return {
        title: (
          <Space>
            <FolderOutlined />
            <Text strong>{category.name}</Text>
            <Tag color="blue">{categoryEndpoints.length} 个接口</Tag>
          </Space>
        ),
        key: `category-${category.id}`,
        children: categoryEndpoints.map(endpoint => ({
          title: (
            <Space>
              <FileTextOutlined />
              <Text>{endpoint.name}</Text>
              <Tag color={getMethodColor(endpoint.method)}>{endpoint.method}</Tag>
              <Text type="secondary" style={{ fontSize: '12px' }}>{endpoint.path}</Text>
            </Space>
          ),
          key: `endpoint-${endpoint.id}`,
          isLeaf: true
        }))
      };
    });
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

  const handleSelect = (selectedKeys: React.Key[], info: any) => {
    setSelectedKeys(selectedKeys);
    if (info.node.isLeaf && selectedKeys[0]) {
      // 从endpoint key中提取ID
      const endpointId = parseInt(selectedKeys[0].toString().replace('endpoint-', ''));
      const endpoint = endpoints.find(ep => ep.id === endpointId);
      setSelectedEndpoint(endpoint || null);
    } else {
      setSelectedEndpoint(null);
    }
  };

  const treeData = buildTreeData();

  return (
    <div style={{ padding: '24px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
        <div>
          <Title level={2} style={{ margin: 0 }}>
            <ApiOutlined /> 接口目录
          </Title>
          <Paragraph type="secondary" style={{ margin: '8px 0 0 0' }}>
            按功能模块分类展示所有API接口，支持快速查找和详细信息查看
          </Paragraph>
        </div>
        <Button 
          icon={<ReloadOutlined />} 
          onClick={loadData}
          loading={loading}
        >
          刷新
        </Button>
      </div>

      <Spin spinning={loading}>
        <Row gutter={[24, 24]}>
          <Col xs={24} lg={8}>
            <Card 
              title={
                <Space>
                  <span>接口分类</span>
                  <Tag color="blue">{categories.length} 个分类</Tag>
                  <Tag color="green">{endpoints.length} 个接口</Tag>
                </Space>
              } 
              size="small"
            >
              <Search
                placeholder="搜索接口..."
                allowClear
                enterButton={<SearchOutlined />}
                value={searchValue}
                onChange={(e) => setSearchValue(e.target.value)}
                style={{ marginBottom: 16 }}
              />
              <Tree
                showLine
                expandedKeys={expandedKeys}
                selectedKeys={selectedKeys}
                onExpand={setExpandedKeys}
                onSelect={handleSelect}
                treeData={treeData}
                height={600}
              />
            </Card>
          </Col>

          <Col xs={24} lg={16}>
            <Card title="接口详情" size="small">
              {selectedEndpoint ? (
                <div>
                  <Space direction="vertical" size="large" style={{ width: '100%' }}>
                    <div>
                      <Title level={4}>
                        <Space>
                          {selectedEndpoint.name}
                          <Tag color={getMethodColor(selectedEndpoint.method)}>
                            {selectedEndpoint.method}
                          </Tag>
                        </Space>
                      </Title>
                      <Text code style={{ fontSize: '14px' }}>{selectedEndpoint.path}</Text>
                    </div>

                    <div>
                      <Text strong>描述：</Text>
                      <Paragraph>{selectedEndpoint.description}</Paragraph>
                    </div>

                    <div>
                      <Text strong>分类：</Text>
                      <Tag color="blue">
                        {categories.find(cat => cat.id === selectedEndpoint.category_id)?.name || '未知分类'}
                      </Tag>
                    </div>

                    <div>
                      <Text strong>来源文件：</Text>
                      <Text code>{selectedEndpoint.source_file}</Text>
                    </div>

                    {selectedEndpoint.request_example && (
                      <div>
                        <Text strong>请求示例：</Text>
                        <pre style={{ 
                          background: '#f5f5f5', 
                          padding: '12px', 
                          borderRadius: '4px',
                          overflow: 'auto',
                          fontSize: '12px'
                        }}>
                          {selectedEndpoint.request_example}
                        </pre>
                      </div>
                    )}

                    {selectedEndpoint.response_example && (
                      <div>
                        <Text strong>响应示例：</Text>
                        <pre style={{ 
                          background: '#f5f5f5', 
                          padding: '12px', 
                          borderRadius: '4px',
                          overflow: 'auto',
                          fontSize: '12px'
                        }}>
                          {selectedEndpoint.response_example}
                        </pre>
                      </div>
                    )}

                    <div>
                      <Text strong>创建时间：</Text>
                      <Text type="secondary">
                        {new Date(selectedEndpoint.created_at).toLocaleString()}
                      </Text>
                    </div>

                    <Divider />

                    <Space>
                      <Button type="primary" icon={<LinkOutlined />}>
                        查看详细文档
                      </Button>
                      <Button icon={<FileTextOutlined />}>
                        测试接口
                      </Button>
                    </Space>
                  </Space>
                </div>
              ) : (
                <div style={{ textAlign: 'center', padding: '60px 0', color: '#999' }}>
                  <ApiOutlined style={{ fontSize: '48px', marginBottom: '16px' }} />
                  <div>请从左侧选择一个接口查看详情</div>
                </div>
              )}
            </Card>
          </Col>
        </Row>
      </Spin>
    </div>
  );
};

export default APICatalog;