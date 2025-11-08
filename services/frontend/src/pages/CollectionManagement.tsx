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
  Select, 
  InputNumber,
  message, 
  Popconfirm,
  Tooltip,
  Descriptions,
  Divider,
  Row,
  Col,
  Alert
} from 'antd';
import { 
  PlusOutlined, 
  EditOutlined, 
  DeleteOutlined,
  InfoCircleOutlined,
  ReloadOutlined,
  RebuildOutlined
} from '@ant-design/icons';
import { collectionApi, Collection } from '../services/taishangApi';

const { Title } = Typography;
const { Option } = Select;
const { TextArea } = Input;

const CollectionManagement: React.FC = () => {
  const [collections, setCollections] = useState<Collection[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [detailModalVisible, setDetailModalVisible] = useState(false);
  const [selectedCollection, setSelectedCollection] = useState<Collection | null>(null);
  const [editingCollection, setEditingCollection] = useState<Collection | null>(null);
  const [form] = Form.useForm();

  // 获取集合列表
  const fetchCollections = async () => {
    setLoading(true);
    try {
      const response = await collectionApi.getAll();
      if (response.success && response.data) {
        // 后端返回的是分页格式，需要提取items数组
        const items = response.data.items || [];
        setCollections(items);
      } else {
        message.error('获取集合列表失败');
      }
    } catch (error) {
      message.error('获取集合列表失败');
      console.error('获取集合列表失败:', error);
    } finally {
      setLoading(false);
    }
  };

  // 初始化加载集合列表
  useEffect(() => {
    fetchCollections();
  }, []);

  // 打开新增集合模态框
  const handleOpenAddModal = () => {
    setEditingCollection(null);
    form.resetFields();
    setModalVisible(true);
  };

  // 打开编辑模态框
  const handleEditCollection = (collection: Collection) => {
    setEditingCollection(collection);
    form.setFieldsValue({
      name: collection.name,
      description: collection.description,
      dims: collection.dims,
      indexType: collection.indexType,
      metricType: collection.metricType,
      // 注意：后端可能没有indexParams和metadata字段，需要处理
      indexParams: collection.extraIndexArgs ? JSON.stringify(collection.extraIndexArgs, null, 2) : '',
      metadata: ''
    });
    setModalVisible(true);
  };

  // 提交表单（新增或更新集合）
  const handleSubmit = async (values: any) => {
    try {
      let extraIndexArgs;
      
      try {
        extraIndexArgs = values.indexParams ? JSON.parse(values.indexParams) : {};
      } catch (e) {
        message.error('索引参数格式不正确，请输入有效的JSON');
        return;
      }

      // 后端API期望的字段格式
      const collectionData = {
        name: values.name,
        description: values.description,
        dims: values.dims,
        indexType: values.indexType,
        metricType: values.metricType,
        extraIndexArgs
      };

      if (editingCollection) {
        // 更新集合
        const response = await collectionApi.update(editingCollection.id, collectionData);
        if (response.success) {
          message.success('集合更新成功');
          setModalVisible(false);
          fetchCollections();
        } else {
          message.error(response.error?.message || '集合更新失败');
        }
      } else {
        // 新增集合
        const response = await collectionApi.create(collectionData);
        if (response.success) {
          message.success('集合创建成功');
          setModalVisible(false);
          form.resetFields();
          fetchCollections();
        } else {
          message.error(response.error?.message || '集合创建失败');
        }
      }
    } catch (error) {
      message.error(editingCollection ? '集合更新失败' : '集合创建失败');
      console.error('操作失败:', error);
    }
  };

  // 删除集合
  const handleDeleteCollection = async (id: string) => {
    try {
      const response = await collectionApi.delete(id);
      if (response.success) {
        message.success('集合删除成功');
        fetchCollections();
      } else {
        message.error(response.error?.message || '集合删除失败');
      }
    } catch (error) {
      message.error('集合删除失败');
      console.error('集合删除失败:', error);
    }
  };

  // 重建集合索引
  const handleRebuildIndex = async (id: string) => {
    try {
      const response = await collectionApi.rebuildIndex(id);
      if (response.success) {
        message.success('索引重建成功');
        fetchCollections();
      } else {
        message.error(response.error?.message || '索引重建失败');
      }
    } catch (error) {
      message.error('索引重建失败');
      console.error('索引重建失败:', error);
    }
  };

  // 查看集合详情
  const handleViewCollectionDetail = async (collection: Collection) => {
    try {
      const response = await collectionApi.get(collection.id);
      if (response.success && response.data) {
        setSelectedCollection(response.data);
        setDetailModalVisible(true);
      } else {
        message.error('获取集合详情失败');
      }
    } catch (error) {
      message.error('获取集合详情失败');
      console.error('获取集合详情失败:', error);
    }
  };

  // 表格列定义
  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 200,
      ellipsis: true,
    },
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '维度',
      dataIndex: 'dims',
      key: 'dims',
      width: 100,
    },
    {
      title: '距离度量',
      dataIndex: 'metricType',
      key: 'metricType',
      width: 120,
      render: (metric: string) => (
        <Tag color="blue">{metric}</Tag>
      ),
    },
    {
      title: '索引类型',
      dataIndex: 'indexType',
      key: 'indexType',
      width: 120,
      render: (indexType: string) => indexType || '-',
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      width: 180,
      render: (date: string) => new Date(date).toLocaleString(),
    },
    {
      title: '操作',
      key: 'action',
      width: 240,
      render: (_: any, record: Collection) => (
        <Space size="small">
          <Tooltip title="查看详情">
            <Button 
              type="link" 
              icon={<InfoCircleOutlined />} 
              onClick={() => handleViewCollectionDetail(record)}
            />
          </Tooltip>
          
          <Tooltip title="编辑">
            <Button 
              type="link" 
              icon={<EditOutlined />} 
              onClick={() => handleOpenEditModal(record)}
            />
          </Tooltip>
          
          <Tooltip title="重建索引">
            <Button 
              type="link" 
              icon={<RebuildOutlined />} 
              onClick={() => handleRebuildIndex(record.id)}
            />
          </Tooltip>
          
          <Popconfirm
            title="确定要删除此集合吗？删除后数据将无法恢复！"
            onConfirm={() => handleDeleteCollection(record.id)}
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

  return (
    <div>
      <Title level={2}>向量集合管理</Title>
      
      <Card>
        <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
          <div>
            <Button 
              type="primary" 
              icon={<PlusOutlined />}
              onClick={handleOpenAddModal}
            >
              创建集合
            </Button>
          </div>
          
          <div>
            <Button 
              icon={<ReloadOutlined />}
              onClick={fetchCollections}
              loading={loading}
            >
              刷新
            </Button>
          </div>
        </div>
        
        <Table
          columns={columns}
          dataSource={collections}
          rowKey="id"
          loading={loading}
          pagination={{
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 个集合`,
          }}
        />
      </Card>

      {/* 新增/编辑集合模态框 */}
      <Modal
        title={editingCollection ? '编辑集合' : '创建集合'}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={null}
        width={800}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{
            indexType: 'hnsw',
            metricType: 'cosine',
            dims: 1536
          }}
        >
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="name"
                label="集合名称"
                rules={[{ required: true, message: '请输入集合名称' }]}
              >
                <Input placeholder="请输入集合名称" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="dims"
                label="向量维度"
                rules={[{ required: true, message: '请输入向量维度' }]}
              >
                <InputNumber 
                  min={1} 
                  max={10000} 
                  placeholder="请输入向量维度" 
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>
          </Row>
          
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="metricType"
                label="距离度量"
                rules={[{ required: true, message: '请选择距离度量' }]}
              >
                <Select placeholder="请选择距离度量">
                  <Option value="cosine">余弦相似度</Option>
                  <Option value="euclidean">欧几里得距离</Option>
                  <Option value="manhattan">曼哈顿距离</Option>
                  <Option value="dotproduct">点积</Option>
                  <Option value="hamming">汉明距离</Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="indexType"
                label="索引类型"
              >
                <Select placeholder="请选择索引类型">
                  <Option value="flat">FLAT</Option>
                  <Option value="ivf_flat">IVF_FLAT</Option>
                  <Option value="ivf_sq8">IVF_SQ8</Option>
                  <Option value="ivf_pq">IVF_PQ</Option>
                  <Option value="hnsw">HNSW</Option>
                  <Option value="annoy">ANNOY</Option>
                  <Option value="rhf_flat">RHF_FLAT</Option>
                  <Option value="rhf_sq8">RHF_SQ8</Option>
                  <Option value="rhf_pq">RHF_PQ</Option>
                  <Option value="sparse_inverted_index">SPARSE_INVERTED_INDEX</Option>
                  <Option value="sparse_wand">SPARSE_WAND</Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>
          
          <Form.Item
            name="description"
            label="描述"
          >
            <TextArea rows={3} placeholder="请输入集合描述" />
          </Form.Item>
          
          <Form.Item
            name="indexParams"
            label="索引参数 (JSON格式)"
          >
            <TextArea 
              rows={4} 
              placeholder='请输入索引参数，例如：{"nlist": 100, "nprobe": 10}' 
            />
          </Form.Item>
          
          <Form.Item
            name="metadata"
            label="元数据 (JSON格式)"
          >
            <TextArea 
              rows={4} 
              placeholder='请输入元数据，例如：{"source": "wikipedia", "category": "science"}' 
            />
          </Form.Item>
          
          <Form.Item>
            <Alert
              message="注意"
              description="集合创建后，向量维度和距离度量将无法修改。请谨慎选择这些参数。"
              type="warning"
              showIcon
              style={{ marginBottom: 16 }}
            />
          </Form.Item>
          
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                {editingCollection ? '更新' : '创建'}
              </Button>
              <Button onClick={() => setModalVisible(false)}>
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* 集合详情模态框 */}
      <Modal
          title="集合详情"
          open={detailModalVisible}
          onCancel={() => setDetailModalVisible(false)}
          footer={[
            <Button key="close" onClick={() => setDetailModalVisible(false)}>
              关闭
            </Button>
          ]}
          width={800}
        >
          {selectedCollection && (
            <div>
              <Descriptions bordered column={2}>
                <Descriptions.Item label="ID">{selectedCollection.id}</Descriptions.Item>
                <Descriptions.Item label="租户ID">{selectedCollection.tenantId}</Descriptions.Item>
                <Descriptions.Item label="名称">{selectedCollection.name}</Descriptions.Item>
                <Descriptions.Item label="描述">{selectedCollection.description || '-'}</Descriptions.Item>
                <Descriptions.Item label="向量维度">{selectedCollection.dims}</Descriptions.Item>
                <Descriptions.Item label="索引类型">{selectedCollection.indexType}</Descriptions.Item>
                <Descriptions.Item label="距离度量">{selectedCollection.metricType}</Descriptions.Item>
                <Descriptions.Item label="模型ID">{selectedCollection.modelId || '-'}</Descriptions.Item>
                <Descriptions.Item label="创建时间" span={2}>
                  {new Date(selectedCollection.createdAt).toLocaleString()}
                </Descriptions.Item>
                <Descriptions.Item label="更新时间" span={2}>
                  {new Date(selectedCollection.updatedAt).toLocaleString()}
                </Descriptions.Item>
                {selectedCollection.extraIndexArgs && (
                  <Descriptions.Item label="索引参数" span={2}>
                    <pre>{JSON.stringify(selectedCollection.extraIndexArgs, null, 2)}</pre>
                  </Descriptions.Item>
                )}
              </Descriptions>
            </div>
          )}
        </Modal>
    </div>
  );
};

export default CollectionManagement;