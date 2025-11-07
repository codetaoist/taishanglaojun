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
  Progress,
  Badge,
  Alert
} from 'antd';
import { 
  PlusOutlined, 
  EditOutlined, 
  DeleteOutlined,
  InfoCircleOutlined,
  ReloadOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  StopOutlined
} from '@ant-design/icons';
import { taskApi, Task, TaskStatus, TaskType } from '../services/taishangApi';

const { Title } = Typography;
const { Option } = Select;
const { TextArea } = Input;

const TaskManagement: React.FC = () => {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [detailModalVisible, setDetailModalVisible] = useState(false);
  const [selectedTask, setSelectedTask] = useState<Task | null>(null);
  const [editingTask, setEditingTask] = useState<Task | null>(null);
  const [statusFilter, setStatusFilter] = useState<TaskStatus | undefined>(undefined);
  const [typeFilter, setTypeFilter] = useState<TaskType | undefined>(undefined);
  const [form] = Form.useForm();

  // 获取任务列表
  const fetchTasks = async () => {
    setLoading(true);
    try {
      const response = await taskApi.getAll(statusFilter, typeFilter);
      if (response.success) {
        setTasks(response.data || []);
      } else {
        message.error('获取任务列表失败');
      }
    } catch (error) {
      message.error('获取任务列表失败');
      console.error('获取任务列表失败:', error);
    } finally {
      setLoading(false);
    }
  };

  // 初始化加载任务列表
  useEffect(() => {
    fetchTasks();
  }, [statusFilter, typeFilter]);

  // 打开新增任务模态框
  const handleOpenAddModal = () => {
    setEditingTask(null);
    form.resetFields();
    setModalVisible(true);
  };

  // 打开编辑任务模态框
  const handleOpenEditModal = (task: Task) => {
    setEditingTask(task);
    form.setFieldsValue({
      name: task.name,
      type: task.type,
      description: task.description,
      config: task.config ? JSON.stringify(task.config, null, 2) : '',
      input: task.input ? JSON.stringify(task.input, null, 2) : ''
    });
    setModalVisible(true);
  };

  // 提交表单（新增或更新任务）
  const handleSubmit = async (values: any) => {
    try {
      let config;
      let input;
      
      try {
        config = values.config ? JSON.parse(values.config) : {};
      } catch (e) {
        message.error('配置格式不正确，请输入有效的JSON');
        return;
      }
      
      try {
        input = values.input ? JSON.parse(values.input) : {};
      } catch (e) {
        message.error('输入格式不正确，请输入有效的JSON');
        return;
      }

      const taskData = {
        name: values.name,
        type: values.type,
        description: values.description,
        config,
        input,
        status: TaskStatus.Pending
      };

      if (editingTask) {
        // 更新任务
        const response = await taskApi.update(editingTask.id, taskData);
        if (response.success) {
          message.success('任务更新成功');
          setModalVisible(false);
          fetchTasks();
        } else {
          message.error(response.message || '任务更新失败');
        }
      } else {
        // 新增任务
        const response = await taskApi.create(taskData);
        if (response.success) {
          message.success('任务创建成功');
          setModalVisible(false);
          form.resetFields();
          fetchTasks();
        } else {
          message.error(response.message || '任务创建失败');
        }
      }
    } catch (error) {
      message.error(editingTask ? '任务更新失败' : '任务创建失败');
      console.error('操作失败:', error);
    }
  };

  // 删除任务
  const handleDeleteTask = async (id: string) => {
    try {
      const response = await taskApi.delete(id);
      if (response.success) {
        message.success('任务删除成功');
        fetchTasks();
      } else {
        message.error(response.message || '任务删除失败');
      }
    } catch (error) {
      message.error('任务删除失败');
      console.error('任务删除失败:', error);
    }
  };

  // 查看任务详情
  const handleViewTaskDetail = async (task: Task) => {
    try {
      const response = await taskApi.get(task.id);
      if (response.success) {
        setSelectedTask(response.data);
        setDetailModalVisible(true);
      } else {
        message.error('获取任务详情失败');
      }
    } catch (error) {
      message.error('获取任务详情失败');
      console.error('获取任务详情失败:', error);
    }
  };

  // 获取任务状态标签
  const getStatusTag = (status: TaskStatus) => {
    const statusMap = {
      [TaskStatus.Pending]: { color: 'default', text: '待处理' },
      [TaskStatus.Running]: { color: 'processing', text: '运行中' },
      [TaskStatus.Completed]: { color: 'success', text: '已完成' },
      [TaskStatus.Failed]: { color: 'error', text: '失败' },
      [TaskStatus.Cancelled]: { color: 'warning', text: '已取消' }
    };
    
    const { color, text } = statusMap[status];
    return <Tag color={color}>{text}</Tag>;
  };

  // 获取任务类型标签
  const getTypeTag = (type: TaskType) => {
    const typeMap = {
      [TaskType.Indexing]: { color: 'blue', text: '索引构建' },
      [TaskType.Training]: { color: 'purple', text: '模型训练' },
      [TaskType.Inference]: { color: 'green', text: '推理' },
      [TaskType.FineTuning]: { color: 'orange', text: '微调' },
      [TaskType.DataProcessing]: { color: 'cyan', text: '数据处理' }
    };
    
    const { color, text } = typeMap[type];
    return <Tag color={color}>{text}</Tag>;
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
      render: (type: TaskType) => getTypeTag(type),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (status: TaskStatus) => getStatusTag(status),
    },
    {
      title: '进度',
      dataIndex: 'progress',
      key: 'progress',
      width: 150,
      render: (progress: number, record: Task) => (
        <div style={{ width: 100 }}>
          {record.status === TaskStatus.Running ? (
            <Progress percent={progress || 0} size="small" />
          ) : (
            <span>{progress || 0}%</span>
          )}
        </div>
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
      render: (_: any, record: Task) => (
        <Space size="small">
          <Tooltip title="查看详情">
            <Button 
              type="link" 
              icon={<InfoCircleOutlined />} 
              onClick={() => handleViewTaskDetail(record)}
            />
          </Tooltip>
          
          {record.status !== TaskStatus.Running && (
            <Tooltip title="编辑">
              <Button 
                type="link" 
                icon={<EditOutlined />} 
                onClick={() => handleOpenEditModal(record)}
              />
            </Tooltip>
          )}
          
          <Popconfirm
            title="确定要删除此任务吗？"
            onConfirm={() => handleDeleteTask(record.id)}
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
      <Title level={2}>任务管理</Title>
      
      <Card>
        <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
          <div>
            <Space>
              <Button 
                type="primary" 
                icon={<PlusOutlined />}
                onClick={handleOpenAddModal}
              >
                创建任务
              </Button>
              
              <Select
                placeholder="筛选状态"
                allowClear
                style={{ width: 150 }}
                value={statusFilter}
                onChange={value => setStatusFilter(value)}
              >
                <Option value={TaskStatus.Pending}>待处理</Option>
                <Option value={TaskStatus.Running}>运行中</Option>
                <Option value={TaskStatus.Completed}>已完成</Option>
                <Option value={TaskStatus.Failed}>失败</Option>
                <Option value={TaskStatus.Cancelled}>已取消</Option>
              </Select>
              
              <Select
                placeholder="筛选类型"
                allowClear
                style={{ width: 150 }}
                value={typeFilter}
                onChange={value => setTypeFilter(value)}
              >
                <Option value={TaskType.Indexing}>索引构建</Option>
                <Option value={TaskType.Training}>模型训练</Option>
                <Option value={TaskType.Inference}>推理</Option>
                <Option value={TaskType.FineTuning}>微调</Option>
                <Option value={TaskType.DataProcessing}>数据处理</Option>
              </Select>
            </Space>
          </div>
          
          <div>
            <Button 
              icon={<ReloadOutlined />}
              onClick={fetchTasks}
              loading={loading}
            >
              刷新
            </Button>
          </div>
        </div>
        
        <Table
          columns={columns}
          dataSource={tasks}
          rowKey="id"
          loading={loading}
          pagination={{
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 个任务`,
          }}
        />
      </Card>

      {/* 新增/编辑任务模态框 */}
      <Modal
        title={editingTask ? '编辑任务' : '创建任务'}
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
                label="任务名称"
                rules={[{ required: true, message: '请输入任务名称' }]}
              >
                <Input placeholder="请输入任务名称" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="type"
                label="任务类型"
                rules={[{ required: true, message: '请选择任务类型' }]}
              >
                <Select placeholder="请选择任务类型">
                  <Option value={TaskType.Indexing}>索引构建</Option>
                  <Option value={TaskType.Training}>模型训练</Option>
                  <Option value={TaskType.Inference}>推理</Option>
                  <Option value={TaskType.FineTuning}>微调</Option>
                  <Option value={TaskType.DataProcessing}>数据处理</Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>
          
          <Form.Item
            name="description"
            label="描述"
          >
            <TextArea rows={3} placeholder="请输入任务描述" />
          </Form.Item>
          
          <Form.Item
            name="config"
            label="配置 (JSON格式)"
          >
            <TextArea 
              rows={4} 
              placeholder='请输入任务配置，例如：{"model": "text-embedding-ada-002", "batch_size": 100}' 
            />
          </Form.Item>
          
          <Form.Item
            name="input"
            label="输入 (JSON格式)"
          >
            <TextArea 
              rows={4} 
              placeholder='请输入任务输入，例如：{"collection_id": "123", "documents": [...]}' 
            />
          </Form.Item>
          
          <Form.Item>
            <Alert
              message="注意"
              description="任务创建后将处于待处理状态，需要手动启动或由系统自动调度执行。"
              type="info"
              showIcon
              style={{ marginBottom: 16 }}
            />
          </Form.Item>
          
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                {editingTask ? '更新' : '创建'}
              </Button>
              <Button onClick={() => setModalVisible(false)}>
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* 任务详情模态框 */}
      <Modal
        title="任务详情"
        open={detailModalVisible}
        onCancel={() => setDetailModalVisible(false)}
        footer={[
          <Button key="close" onClick={() => setDetailModalVisible(false)}>
            关闭
          </Button>
        ]}
        width={900}
      >
        {selectedTask && (
          <div>
            <Descriptions bordered column={2}>
              <Descriptions.Item label="ID">{selectedTask.id}</Descriptions.Item>
              <Descriptions.Item label="名称">{selectedTask.name}</Descriptions.Item>
              <Descriptions.Item label="类型">{getTypeTag(selectedTask.type)}</Descriptions.Item>
              <Descriptions.Item label="状态">{getStatusTag(selectedTask.status)}</Descriptions.Item>
              <Descriptions.Item label="进度">
                <Progress percent={selectedTask.progress || 0} size="small" />
              </Descriptions.Item>
              <Descriptions.Item label="创建时间">
                {new Date(selectedTask.createdAt).toLocaleString()}
              </Descriptions.Item>
              <Descriptions.Item label="更新时间">
                {new Date(selectedTask.updatedAt).toLocaleString()}
              </Descriptions.Item>
              <Descriptions.Item label="开始时间">
                {selectedTask.startedAt ? new Date(selectedTask.startedAt).toLocaleString() : '-'}
              </Descriptions.Item>
              <Descriptions.Item label="完成时间">
                {selectedTask.completedAt ? new Date(selectedTask.completedAt).toLocaleString() : '-'}
              </Descriptions.Item>
              <Descriptions.Item label="描述" span={2}>
                {selectedTask.description || '-'}
              </Descriptions.Item>
              {selectedTask.error && (
                <Descriptions.Item label="错误信息" span={2}>
                  <Alert message={selectedTask.error} type="error" />
                </Descriptions.Item>
              )}
            </Descriptions>
            
            {selectedTask.config && (
              <>
                <Divider>配置</Divider>
                <Row>
                  <Col span={24}>
                    <pre style={{ 
                      background: '#f5f5f5', 
                      padding: '10px', 
                      borderRadius: '4px',
                      overflow: 'auto',
                      maxHeight: '200px'
                    }}>
                      {JSON.stringify(selectedTask.config, null, 2)}
                    </pre>
                  </Col>
                </Row>
              </>
            )}
            
            {selectedTask.input && (
              <>
                <Divider>输入</Divider>
                <Row>
                  <Col span={24}>
                    <pre style={{ 
                      background: '#f5f5f5', 
                      padding: '10px', 
                      borderRadius: '4px',
                      overflow: 'auto',
                      maxHeight: '200px'
                    }}>
                      {JSON.stringify(selectedTask.input, null, 2)}
                    </pre>
                  </Col>
                </Row>
              </>
            )}
            
            {selectedTask.output && (
              <>
                <Divider>输出</Divider>
                <Row>
                  <Col span={24}>
                    <pre style={{ 
                      background: '#f5f5f5', 
                      padding: '10px', 
                      borderRadius: '4px',
                      overflow: 'auto',
                      maxHeight: '200px'
                    }}>
                      {JSON.stringify(selectedTask.output, null, 2)}
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

export default TaskManagement;