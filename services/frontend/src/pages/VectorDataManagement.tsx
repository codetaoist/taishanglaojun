import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Typography, 
  Table, 
  Button, 
  Space, 
  Tag, 
  Modal, 
  Form, 
  Input, 
  InputNumber,
  message, 
  Popconfirm,
  Tooltip,
  Descriptions,
  Divider,
  Row,
  Col,
  Alert,
  Tabs,
  Select,
  Upload,
  Badge
} from 'antd';
import { 
  PlusOutlined, 
  EditOutlined, 
  DeleteOutlined,
  InfoCircleOutlined,
  ReloadOutlined,
  SearchOutlined,
  UploadOutlined,
  DownloadOutlined,
  FileTextOutlined
} from '@ant-design/icons';
import { collectionApi } from '../services/taishangApi';
import { vectorDataApi } from '../services/vectorDataApi';
import type { Collection } from '../services/taishangApi';
import type { VectorData, VectorSearchRequest } from '../services/vectorDataApi';

const { Title } = Typography;
const { Option } = Select;
const { TextArea } = Input;
const { TabPane } = Tabs;

const VectorDataManagement: React.FC = () => {
  const [collections, setCollections] = useState<Collection[]>([]);
  const [selectedCollection, setSelectedCollection] = useState<Collection | null>(null);
  const [vectorData, setVectorData] = useState<VectorData[]>([]);
  const [searchResults, setSearchResults] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchLoading, setSearchLoading] = useState(false);
  const [vectorModalVisible, setVectorModalVisible] = useState(false);
  const [searchModalVisible, setSearchModalVisible] = useState(false);
  const [detailModalVisible, setDetailModalVisible] = useState(false);
  const [selectedVector, setSelectedVector] = useState<VectorData | null>(null);
  const [editingVector, setEditingVector] = useState<VectorData | null>(null);
  const [vectorForm] = Form.useForm();
  const [searchForm] = Form.useForm();
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0
  });

  // 获取集合列表
  const fetchCollections = async () => {
    try {
      const response = await collectionApi.getAll();
      if (response.code === 200 && response.data) {
        const items = response.data.items || [];
        setCollections(items);
        if (items.length > 0 && !selectedCollection) {
          setSelectedCollection(items[0]);
        }
      } else {
        message.error('获取集合列表失败');
      }
    } catch (error) {
      message.error('获取集合列表失败');
      console.error('获取集合列表失败:', error);
    }
  };

  // 获取向量数据列表
  const fetchVectorData = async (page = 1, pageSize = 10) => {
    if (!selectedCollection) return;
    
    setLoading(true);
    try {
      const response = await vectorDataApi.getAll(selectedCollection.name, page, pageSize);
      if (response.code === 200 && response.data) {
        setVectorData(response.data.items || []);
        setPagination({
          current: page,
          pageSize: pageSize,
          total: response.data.total || 0
        });
      } else {
        message.error('获取向量数据失败');
      }
    } catch (error) {
      message.error('获取向量数据失败');
      console.error('获取向量数据失败:', error);
    } finally {
      setLoading(false);
    }
  };

  // 初始化加载
  useEffect(() => {
    fetchCollections();
  }, []);

  // 当选择的集合变化时，加载对应的向量数据
  useEffect(() => {
    if (selectedCollection) {
      fetchVectorData(1, pagination.pageSize);
    }
  }, [selectedCollection]);

  // 处理表格分页变化
  const handleTableChange = (paginationInfo: any) => {
    fetchVectorData(paginationInfo.current, paginationInfo.pageSize);
  };

  // 打开新增向量模态框
  const handleOpenAddVectorModal = () => {
    setEditingVector(null);
    vectorForm.resetFields();
    setVectorModalVisible(true);
  };

  // 打开编辑向量模态框
  const handleOpenEditVectorModal = (vector: VectorData) => {
    setEditingVector(vector);
    vectorForm.setFieldsValue({
      id: vector.id,
      vector: vector.vector.join(', '),
      metadata: vector.metadata ? JSON.stringify(vector.metadata, null, 2) : ''
    });
    setVectorModalVisible(true);
  };

  // 提交向量表单
  const handleVectorSubmit = async (values: any) => {
    if (!selectedCollection) {
      message.error('请先选择一个集合');
      return;
    }

    try {
      let vectorArray: number[] = [];
      let metadata: Record<string, any> = {};
      
      // 解析向量数据
      try {
        vectorArray = values.vector.split(',').map((v: string) => parseFloat(v.trim()));
        if (vectorArray.some(isNaN)) {
          throw new Error('向量数据包含非数字值');
        }
      } catch (e) {
        message.error('向量数据格式不正确，请输入逗号分隔的数字');
        return;
      }
      
      // 解析元数据
      try {
        metadata = values.metadata ? JSON.parse(values.metadata) : {};
      } catch (e) {
        message.error('元数据格式不正确，请输入有效的JSON');
        return;
      }

      const vectorData: VectorData = {
        id: values.id,
        vector: vectorArray,
        metadata
      };

      const response = await vectorDataApi.upsert(selectedCollection.name, [vectorData]);
      if (response.code === 200) {
        message.success('向量数据保存成功');
        setVectorModalVisible(false);
        vectorForm.resetFields();
        fetchVectorData(pagination.current, pagination.pageSize);
      } else {
        message.error(response.message || '向量数据保存失败');
      }
    } catch (error) {
      message.error('向量数据保存失败');
      console.error('向量数据保存失败:', error);
    }
  };

  // 删除向量数据
  const handleDeleteVector = async (vectorId: string) => {
    if (!selectedCollection) return;
    
    try {
      const response = await vectorDataApi.delete(selectedCollection.name, [vectorId]);
      if (response.code === 200) {
        message.success('向量数据删除成功');
        fetchVectorData(pagination.current, pagination.pageSize);
      } else {
        message.error(response.message || '向量数据删除失败');
      }
    } catch (error) {
      message.error('向量数据删除失败');
      console.error('向量数据删除失败:', error);
    }
  };

  // 查看向量详情
  const handleViewVectorDetail = (vector: VectorData) => {
    setSelectedVector(vector);
    setDetailModalVisible(true);
  };

  // 打开搜索模态框
  const handleOpenSearchModal = () => {
    searchForm.resetFields();
    setSearchResults([]);
    setSearchModalVisible(true);
  };

  // 执行向量搜索
  const handleSearch = async (values: any) => {
    if (!selectedCollection) {
      message.error('请先选择一个集合');
      return;
    }

    try {
      let queryVector: number[] = [];
      
      // 解析查询向量
      try {
        queryVector = values.queryVector.split(',').map((v: string) => parseFloat(v.trim()));
        if (queryVector.some(isNaN)) {
          throw new Error('查询向量包含非数字值');
        }
      } catch (e) {
        message.error('查询向量格式不正确，请输入逗号分隔的数字');
        return;
      }

      setSearchLoading(true);
      const searchRequest: VectorSearchRequest = {
        collectionName: selectedCollection.name,
        vector: queryVector,
        topK: values.topK || 10
      };

      const response = await vectorDataApi.search(searchRequest);
      if (response.code === 200 && response.data) {
        setSearchResults(response.data.results || []);
        message.success(`找到 ${response.data.results?.length || 0} 个相似向量`);
      } else {
        message.error(response.message || '向量搜索失败');
      }
    } catch (error) {
      message.error('向量搜索失败');
      console.error('向量搜索失败:', error);
    } finally {
      setSearchLoading(false);
    }
  };

  // 向量数据表格列定义
  const vectorColumns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 200,
      ellipsis: true,
    },
    {
      title: '向量维度',
      key: 'dimensions',
      width: 120,
      render: (_: any, record: VectorData) => (
        <span>{record.vector.length} 维</span>
      ),
    },
    {
      title: '元数据',
      dataIndex: 'metadata',
      key: 'metadata',
      ellipsis: true,
      render: (metadata: Record<string, any>) => (
        <span>{metadata ? Object.keys(metadata).join(', ') : '-'}</span>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      width: 180,
      render: (date: string) => date ? new Date(date).toLocaleString() : '-',
    },
    {
      title: '操作',
      key: 'action',
      width: 180,
      render: (_: any, record: VectorData) => (
        <Space size="small">
          <Tooltip title="查看详情">
            <Button 
              type="link" 
              icon={<InfoCircleOutlined />} 
              onClick={() => handleViewVectorDetail(record)}
            />
          </Tooltip>
          
          <Tooltip title="编辑">
            <Button 
              type="link" 
              icon={<EditOutlined />} 
              onClick={() => handleOpenEditVectorModal(record)}
            />
          </Tooltip>
          
          <Popconfirm
            title="确定要删除此向量数据吗？"
            onConfirm={() => handleDeleteVector(record.id)}
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
      ),
    },
  ];

  // 搜索结果表格列定义
  const searchResultColumns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 200,
      ellipsis: true,
    },
    {
      title: '相似度分数',
      dataIndex: 'score',
      key: 'score',
      width: 120,
      render: (score: number) => (
        <Tag color={score > 0.8 ? 'green' : score > 0.5 ? 'orange' : 'red'}>
          {score.toFixed(4)}
        </Tag>
      ),
    },
    {
      title: '元数据',
      dataIndex: 'metadata',
      key: 'metadata',
      ellipsis: true,
      render: (metadata: Record<string, any>) => (
        <span>{metadata ? Object.keys(metadata).join(', ') : '-'}</span>
      ),
    },
  ];

  return (
    <div>
      <Title level={2}>向量数据管理</Title>
      
      <Card>
        <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
          <div>
            <Select
              style={{ width: 300, marginRight: 16 }}
              placeholder="请选择向量集合"
              value={selectedCollection?.name}
              onChange={(name) => {
                const collection = collections.find(c => c.name === name);
                setSelectedCollection(collection || null);
              }}
            >
              {collections.map(collection => (
                <Option key={collection.name} value={collection.name}>
                  {collection.name} ({collection.dims}维)
                </Option>
              ))}
            </Select>
            
            {selectedCollection && (
              <>
                <Button 
                  type="primary" 
                  icon={<PlusOutlined />}
                  onClick={handleOpenAddVectorModal}
                >
                  添加向量
                </Button>
                
                <Button 
                  style={{ marginLeft: 8 }}
                  icon={<SearchOutlined />}
                  onClick={handleOpenSearchModal}
                >
                  向量搜索
                </Button>
              </>
            )}
          </div>
          
          <div>
            <Button 
              icon={<ReloadOutlined />}
              onClick={() => fetchVectorData(pagination.current, pagination.pageSize)}
              loading={loading}
            >
              刷新
            </Button>
          </div>
        </div>
        
        {selectedCollection ? (
          <Table
            columns={vectorColumns}
            dataSource={vectorData}
            rowKey="id"
            loading={loading}
            pagination={{
              ...pagination,
              showSizeChanger: true,
              showQuickJumper: true,
              showTotal: (total) => `共 ${total} 个向量`,
            }}
            onChange={handleTableChange}
          />
        ) : (
          <div style={{ textAlign: 'center', padding: '50px 0', color: '#999' }}>
            请选择一个向量集合以查看向量数据
          </div>
        )}
      </Card>

      {/* 添加/编辑向量模态框 */}
      <Modal
        title={editingVector ? '编辑向量' : '添加向量'}
        open={vectorModalVisible}
        onCancel={() => setVectorModalVisible(false)}
        footer={null}
        width={800}
      >
        <Form
          form={vectorForm}
          layout="vertical"
          onFinish={handleVectorSubmit}
        >
          <Form.Item
            name="id"
            label="向量ID"
            rules={[{ required: true, message: '请输入向量ID' }]}
          >
            <Input placeholder="请输入向量ID" />
          </Form.Item>
          
          <Form.Item
            name="vector"
            label="向量数据"
            rules={[{ required: true, message: '请输入向量数据' }]}
            extra="请输入逗号分隔的数字，例如: 0.1, 0.2, 0.3, ..."
          >
            <TextArea 
              rows={4} 
              placeholder="请输入向量数据，例如: 0.1, 0.2, 0.3, ..." 
            />
          </Form.Item>
          
          <Form.Item
            name="metadata"
            label="元数据 (JSON格式)"
          >
            <TextArea 
              rows={4} 
              placeholder='请输入元数据，例如：{"title": "文档标题", "category": "分类"}' 
            />
          </Form.Item>
          
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                {editingVector ? '更新' : '添加'}
              </Button>
              <Button onClick={() => setVectorModalVisible(false)}>
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* 向量搜索模态框 */}
      <Modal
        title="向量搜索"
        open={searchModalVisible}
        onCancel={() => setSearchModalVisible(false)}
        footer={null}
        width={800}
      >
        <Form
          form={searchForm}
          layout="vertical"
          onFinish={handleSearch}
        >
          <Form.Item
            name="queryVector"
            label="查询向量"
            rules={[{ required: true, message: '请输入查询向量' }]}
            extra="请输入逗号分隔的数字，例如: 0.1, 0.2, 0.3, ..."
          >
            <TextArea 
              rows={4} 
              placeholder="请输入查询向量，例如: 0.1, 0.2, 0.3, ..." 
            />
          </Form.Item>
          
          <Form.Item
            name="topK"
            label="返回结果数量"
            initialValue={10}
          >
            <InputNumber 
              min={1} 
              max={100} 
              placeholder="请输入返回结果数量" 
              style={{ width: '100%' }}
            />
          </Form.Item>
          
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={searchLoading}>
                搜索
              </Button>
              <Button onClick={() => setSearchModalVisible(false)}>
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
        
        {searchResults.length > 0 && (
          <div style={{ marginTop: 24 }}>
            <Divider>搜索结果</Divider>
            <Table
              columns={searchResultColumns}
              dataSource={searchResults}
              rowKey="id"
              pagination={false}
              size="small"
            />
          </div>
        )}
      </Modal>

      {/* 向量详情模态框 */}
      <Modal
        title="向量详情"
        open={detailModalVisible}
        onCancel={() => setDetailModalVisible(false)}
        footer={[
          <Button key="close" onClick={() => setDetailModalVisible(false)}>
            关闭
          </Button>
        ]}
        width={800}
      >
        {selectedVector && (
          <div>
            <Descriptions bordered column={2}>
              <Descriptions.Item label="ID">{selectedVector.id}</Descriptions.Item>
              <Descriptions.Item label="向量维度">{selectedVector.vector.length} 维</Descriptions.Item>
              <Descriptions.Item label="创建时间" span={2}>
                {selectedVector.createdAt ? new Date(selectedVector.createdAt).toLocaleString() : '-'}
              </Descriptions.Item>
              <Descriptions.Item label="向量数据" span={2}>
                <div style={{ maxHeight: '200px', overflow: 'auto' }}>
                  <pre>{selectedVector.vector.slice(0, 10).join(', ')}{selectedVector.vector.length > 10 ? '...' : ''}</pre>
                </div>
              </Descriptions.Item>
              {selectedVector.metadata && (
                <Descriptions.Item label="元数据" span={2}>
                  <pre>{JSON.stringify(selectedVector.metadata, null, 2)}</pre>
                </Descriptions.Item>
              )}
            </Descriptions>
          </div>
        )}
      </Modal>
    </div>
  );
};

export default VectorDataManagement;