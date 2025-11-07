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
  message, 
  Popconfirm,
  Tooltip,
  Descriptions,
  Divider,
  Row,
  Col,
  Switch
} from 'antd';
import { 
  PlusOutlined, 
  EditOutlined, 
  DeleteOutlined,
  InfoCircleOutlined,
  ReloadOutlined
} from '@ant-design/icons';
import { modelApi, Model } from '../services/taishangApi';

const { Title } = Typography;
const { Option } = Select;
const { TextArea } = Input;

const ModelManagement: React.FC = () => {
  const [models, setModels] = useState<Model[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [detailModalVisible, setDetailModalVisible] = useState(false);
  const [selectedModel, setSelectedModel] = useState<Model | null>(null);
  const [editingModel, setEditingModel] = useState<Model | null>(null);
  const [form] = Form.useForm();

  // 获取模型列表
  const fetchModels = async () => {
    setLoading(true);
    try {
      const response = await modelApi.getAll();
      if (response.success) {
        setModels(response.data || []);
      } else {
        message.error('获取模型列表失败');
      }
    } catch (error) {
      message.error('获取模型列表失败');
      console.error('获取模型列表失败:', error);
    } finally {
      setLoading(false);
    }
  };

  // 初始化加载模型列表
  useEffect(() => {
    fetchModels();
  }, []);

  // 打开新增模型模态框
  const handleOpenAddModal = () => {
    setEditingModel(null);
    form.resetFields();
    setModalVisible(true);
  };

  // 打开编辑模型模态框
  const handleOpenEditModal = (model: Model) => {
    setEditingModel(model);
    form.setFieldsValue({
      name: model.name,
      type: model.type,
      provider: model.provider,
      version: model.version,
      description: model.description,
      config: model.config ? JSON.stringify(model.config, null, 2) : '',
      capabilities: model.capabilities || []
    });
    setModalVisible(true);
  };

  // 提交表单（新增或更新模型）
  const handleSubmit = async (values: any) => {
    try {
      let config;
      try {
        config = values.config ? JSON.parse(values.config) : {};
      } catch (e) {
        message.error('配置格式不正确，请输入有效的JSON');
        return;
      }

      const modelData = {
        name: values.name,
        type: values.type,
        provider: values.provider,
        version: values.version,
        description: values.description,
        config,
        capabilities: values.capabilities || []
      };

      if (editingModel) {
        // 更新模型
        const response = await modelApi.update(editingModel.id, modelData);
        if (response.success) {
          message.success('模型更新成功');
          setModalVisible(false);
          fetchModels();
        } else {
          message.error(response.message || '模型更新失败');
        }
      } else {
        // 新增模型
        const response = await modelApi.register(modelData);
        if (response.success) {
          message.success('模型注册成功');
          setModalVisible(false);
          form.resetFields();
          fetchModels();
        } else {
          message.error(response.message || '模型注册失败');
        }
      }
    } catch (error) {
      message.error(editingModel ? '模型更新失败' : '模型注册失败');
      console.error('操作失败:', error);
    }
  };

  // 删除模型
  const handleDeleteModel = async (id: string) => {
    try {
      const response = await modelApi.delete(id);
      if (response.success) {
        message.success('模型删除成功');
        fetchModels();
      } else {
        message.error(response.message || '模型删除失败');
      }
    } catch (error) {
      message.error('模型删除失败');
      console.error('模型删除失败:', error);
    }
  };

  // 查看模型详情
  const handleViewModelDetail = async (model: Model) => {
    try {
      const response = await modelApi.get(model.id);
      if (response.success) {
        setSelectedModel(response.data);
        setDetailModalVisible(true);
      } else {
        message.error('获取模型详情失败');
      }
    } catch (error) {
      message.error('获取模型详情失败');
      console.error('获取模型详情失败:', error);
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
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      width: 120,
      render: (type: string) => (
        <Tag color="blue">{type}</Tag>
      ),
    },
    {
      title: '提供商',
      dataIndex: 'provider',
      key: 'provider',
      width: 120,
      render: (provider: string) => (
        <Tag color="green">{provider}</Tag>
      ),
    },
    {
      title: '版本',
      dataIndex: 'version',
      key: 'version',
      width: 100,
      render: (version: string) => version || '-',
    },
    {
      title: '能力',
      dataIndex: 'capabilities',
      key: 'capabilities',
      width: 200,
      render: (capabilities: string[]) => (
        <>
          {capabilities && capabilities.length > 0 ? (
            capabilities.slice(0, 2).map(capability => (
              <Tag key={capability} color="purple" style={{ marginBottom: 4 }}>
                {capability}
              </Tag>
            ))
          ) : (
            '-'
          )}
          {capabilities && capabilities.length > 2 && (
            <Tooltip title={capabilities.slice(2).join(', ')}>
              <Tag color="purple">+{capabilities.length - 2}</Tag>
            </Tooltip>
          )}
        </>
      ),
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
      width: 200,
      render: (_: any, record: Model) => (
        <Space size="small">
          <Tooltip title="查看详情">
            <Button 
              type="link" 
              icon={<InfoCircleOutlined />} 
              onClick={() => handleViewModelDetail(record)}
            />
          </Tooltip>
          
          <Tooltip title="编辑">
            <Button 
              type="link" 
              icon={<EditOutlined />} 
              onClick={() => handleOpenEditModal(record)}
            />
          </Tooltip>
          
          <Popconfirm
            title="确定要删除此模型吗？"
            onConfirm={() => handleDeleteModel(record.id)}
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
      <Title level={2}>模型管理</Title>
      
      <Card>
        <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
          <div>
            <Button 
              type="primary" 
              icon={<PlusOutlined />}
              onClick={handleOpenAddModal}
            >
              注册模型
            </Button>
          </div>
          
          <div>
            <Button 
              icon={<ReloadOutlined />}
              onClick={fetchModels}
              loading={loading}
            >
              刷新
            </Button>
          </div>
        </div>
        
        <Table
          columns={columns}
          dataSource={models}
          rowKey="id"
          loading={loading}
          pagination={{
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 个模型`,
          }}
        />
      </Card>

      {/* 新增/编辑模型模态框 */}
      <Modal
        title={editingModel ? '编辑模型' : '注册模型'}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={null}
        width={800}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
        >
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="name"
                label="模型名称"
                rules={[{ required: true, message: '请输入模型名称' }]}
              >
                <Input placeholder="请输入模型名称" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="type"
                label="模型类型"
                rules={[{ required: true, message: '请选择模型类型' }]}
              >
                <Select placeholder="请选择模型类型">
                  <Option value="llm">大语言模型</Option>
                  <Option value="embedding">嵌入模型</Option>
                  <Option value="rerank">重排序模型</Option>
                  <Option value="image">图像模型</Option>
                  <Option value="audio">音频模型</Option>
                  <Option value="multimodal">多模态模型</Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>
          
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="provider"
                label="提供商"
                rules={[{ required: true, message: '请选择提供商' }]}
              >
                <Select placeholder="请选择提供商">
                  <Option value="openai">OpenAI</Option>
                  <Option value="anthropic">Anthropic</Option>
                  <Option value="huggingface">Hugging Face</Option>
                  <Option value="cohere">Cohere</Option>
                  <Option value="azure">Azure</Option>
                  <Option value="local">本地</Option>
                  <Option value="custom">自定义</Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="version"
                label="版本"
              >
                <Input placeholder="请输入版本号" />
              </Form.Item>
            </Col>
          </Row>
          
          <Form.Item
            name="description"
            label="描述"
          >
            <TextArea rows={3} placeholder="请输入模型描述" />
          </Form.Item>
          
          <Form.Item
            name="capabilities"
            label="能力"
          >
            <Select mode="multiple" placeholder="请选择模型能力">
              <Option value="text-generation">文本生成</Option>
              <Option value="text-embedding">文本嵌入</Option>
              <Option value="text-classification">文本分类</Option>
              <Option value="text-summarization">文本摘要</Option>
              <Option value="question-answering">问答</Option>
              <Option value="translation">翻译</Option>
              <Option value="code-generation">代码生成</Option>
              <Option value="image-generation">图像生成</Option>
              <Option value="image-recognition">图像识别</Option>
              <Option value="speech-to-text">语音转文本</Option>
              <Option value="text-to-speech">文本转语音</Option>
            </Select>
          </Form.Item>
          
          <Form.Item
            name="config"
            label="配置 (JSON格式)"
          >
            <TextArea 
              rows={6} 
              placeholder='请输入模型配置，例如：{"api_key": "your_key", "temperature": 0.7}' 
            />
          </Form.Item>
          
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                {editingModel ? '更新' : '注册'}
              </Button>
              <Button onClick={() => setModalVisible(false)}>
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* 模型详情模态框 */}
      <Modal
        title="模型详情"
        open={detailModalVisible}
        onCancel={() => setDetailModalVisible(false)}
        footer={[
          <Button key="close" onClick={() => setDetailModalVisible(false)}>
            关闭
          </Button>
        ]}
        width={800}
      >
        {selectedModel && (
          <div>
            <Descriptions bordered column={2}>
              <Descriptions.Item label="ID">{selectedModel.id}</Descriptions.Item>
              <Descriptions.Item label="名称">{selectedModel.name}</Descriptions.Item>
              <Descriptions.Item label="类型">{selectedModel.type}</Descriptions.Item>
              <Descriptions.Item label="提供商">{selectedModel.provider}</Descriptions.Item>
              <Descriptions.Item label="版本">{selectedModel.version || '-'}</Descriptions.Item>
              <Descriptions.Item label="创建时间">
                {new Date(selectedModel.createdAt).toLocaleString()}
              </Descriptions.Item>
              <Descriptions.Item label="更新时间">
                {new Date(selectedModel.updatedAt).toLocaleString()}
              </Descriptions.Item>
              <Descriptions.Item label="描述" span={2}>
                {selectedModel.description || '-'}
              </Descriptions.Item>
              <Descriptions.Item label="能力" span={2}>
                {selectedModel.capabilities && selectedModel.capabilities.length > 0 ? (
                  <>
                    {selectedModel.capabilities.map(capability => (
                      <Tag key={capability} color="purple" style={{ marginBottom: 4 }}>
                        {capability}
                      </Tag>
                    ))}
                  </>
                ) : '-'}
              </Descriptions.Item>
            </Descriptions>
            
            {selectedModel.config && (
              <>
                <Divider>配置</Divider>
                <Row>
                  <Col span={24}>
                    <pre style={{ 
                      background: '#f5f5f5', 
                      padding: '10px', 
                      borderRadius: '4px',
                      overflow: 'auto',
                      maxHeight: '300px'
                    }}>
                      {JSON.stringify(selectedModel.config, null, 2)}
                    </pre>
                  </Col>
                </Row>
              </>
            )}
          </div>
        )}
      </Modal>
    </div>
  );
};

export default ModelManagement;