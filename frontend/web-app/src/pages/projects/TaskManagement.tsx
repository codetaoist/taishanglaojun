import React, { useState, useEffect } from 'react';
import {
  Card,
  Row,
  Col,
  Button,
  Table,
  Tag,
  Progress,
  Space,
  Avatar,
  Tooltip,
  Modal,
  Form,
  Input,
  Select,
  DatePicker,
  message,
  Dropdown,
  Menu,
  Tabs,
  List,
  Checkbox,
  Badge
} from 'antd';
import {
  PlusOutlined,
  CheckSquareOutlined,
  ClockCircleOutlined,
  UserOutlined,
  CalendarOutlined,
  FlagOutlined,
  MoreOutlined,
  EditOutlined,
  DeleteOutlined,
  EyeOutlined,
  FilterOutlined
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';

const { Option } = Select;
const { TextArea } = Input;
const { TabPane } = Tabs;

interface Task {
  id: string;
  title: string;
  description: string;
  status: 'todo' | 'in_progress' | 'review' | 'completed';
  priority: 'low' | 'medium' | 'high' | 'urgent';
  assignee: {
    id: string;
    name: string;
    avatar?: string;
  };
  project: {
    id: string;
    name: string;
  };
  dueDate: string;
  createdDate: string;
  completedDate?: string;
  estimatedHours: number;
  actualHours?: number;
  tags: string[];
  subtasks?: Array<{
    id: string;
    title: string;
    completed: boolean;
  }>;
}

const TaskManagement: React.FC = () => {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(false);
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [activeTab, setActiveTab] = useState('all');
  const [form] = Form.useForm();

  // 模拟数据
  useEffect(() => {
    setLoading(true);
    setTimeout(() => {
      setTasks([
        {
          id: '1',
          title: '完成文献资料整理',
          description: '整理太上老君相关的古代文献资料，建立分类索引',
          status: 'in_progress',
          priority: 'high',
          assignee: { id: '1', name: '张三' },
          project: { id: '1', name: '太上老君文化研究项目' },
          dueDate: '2024-02-15',
          createdDate: '2024-01-20',
          estimatedHours: 16,
          actualHours: 8,
          tags: ['文献', '整理', '分类'],
          subtasks: [
            { id: '1-1', title: '收集古代文献', completed: true },
            { id: '1-2', title: '建立分类体系', completed: true },
            { id: '1-3', title: '录入数据库', completed: false }
          ]
        },
        {
          id: '2',
          title: '撰写研究报告第一章',
          description: '完成太上老君思想体系概述章节的撰写',
          status: 'todo',
          priority: 'medium',
          assignee: { id: '2', name: '李四' },
          project: { id: '1', name: '太上老君文化研究项目' },
          dueDate: '2024-02-20',
          createdDate: '2024-01-22',
          estimatedHours: 24,
          tags: ['写作', '研究', '报告']
        },
        {
          id: '3',
          title: '设计学习平台UI界面',
          description: '设计智慧学习平台的用户界面，包括主页、课程页面等',
          status: 'review',
          priority: 'high',
          assignee: { id: '3', name: '王五' },
          project: { id: '3', name: '智慧学习平台开发' },
          dueDate: '2024-02-10',
          createdDate: '2024-01-15',
          estimatedHours: 32,
          actualHours: 28,
          tags: ['设计', 'UI', '界面']
        },
        {
          id: '4',
          title: '开发用户认证模块',
          description: '实现用户注册、登录、权限管理等功能',
          status: 'completed',
          priority: 'urgent',
          assignee: { id: '4', name: '赵六' },
          project: { id: '3', name: '智慧学习平台开发' },
          dueDate: '2024-01-30',
          createdDate: '2024-01-10',
          completedDate: '2024-01-28',
          estimatedHours: 20,
          actualHours: 18,
          tags: ['开发', '认证', '安全']
        }
      ]);
      setLoading(false);
    }, 1000);
  }, []);

  const getStatusColor = (status: Task['status']) => {
    const colors = {
      todo: 'default',
      in_progress: 'processing',
      review: 'warning',
      completed: 'success'
    };
    return colors[status];
  };

  const getStatusText = (status: Task['status']) => {
    const texts = {
      todo: '待开始',
      in_progress: '进行中',
      review: '待审核',
      completed: '已完成'
    };
    return texts[status];
  };

  const getPriorityColor = (priority: Task['priority']) => {
    const colors = {
      low: 'green',
      medium: 'blue',
      high: 'orange',
      urgent: 'red'
    };
    return colors[priority];
  };

  const getPriorityText = (priority: Task['priority']) => {
    const texts = {
      low: '低',
      medium: '中',
      high: '高',
      urgent: '紧急'
    };
    return texts[priority];
  };

  const handleCreateTask = async (values: any) => {
    try {
      setLoading(true);
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      const newTask: Task = {
        id: Date.now().toString(),
        title: values.title,
        description: values.description,
        status: 'todo',
        priority: values.priority,
        assignee: { id: '1', name: '当前用户' },
        project: { id: values.project, name: '选中的项目' },
        dueDate: values.dueDate.format('YYYY-MM-DD'),
        createdDate: dayjs().format('YYYY-MM-DD'),
        estimatedHours: values.estimatedHours,
        tags: values.tags || []
      };

      setTasks([newTask, ...tasks]);
      setCreateModalVisible(false);
      form.resetFields();
      message.success('任务创建成功！');
    } catch (error) {
      message.error('任务创建失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  const handleStatusChange = (taskId: string, newStatus: Task['status']) => {
    setTasks(tasks.map(task => 
      task.id === taskId 
        ? { 
            ...task, 
            status: newStatus,
            completedDate: newStatus === 'completed' ? dayjs().format('YYYY-MM-DD') : undefined
          }
        : task
    ));
    message.success('任务状态更新成功');
  };

  const handleMenuClick = (key: string, task: Task) => {
    switch (key) {
      case 'view':
        message.info(`查看任务: ${task.title}`);
        break;
      case 'edit':
        message.info(`编辑任务: ${task.title}`);
        break;
      case 'delete':
        Modal.confirm({
          title: '确认删除',
          content: `确定要删除任务"${task.title}"吗？`,
          onOk: () => {
            setTasks(tasks.filter(t => t.id !== task.id));
            message.success('任务删除成功');
          }
        });
        break;
    }
  };

  const columns: ColumnsType<Task> = [
    {
      title: '任务标题',
      dataIndex: 'title',
      key: 'title',
      render: (text, record) => (
        <Space direction="vertical" size={0}>
          <span style={{ fontWeight: 500 }}>{text}</span>
          <Space size={4}>
            {record.tags.map(tag => (
              <Tag key={tag} size="small">{tag}</Tag>
            ))}
          </Space>
        </Space>
      )
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status, record) => (
        <Select
          value={status}
          size="small"
          style={{ width: 100 }}
          onChange={(value) => handleStatusChange(record.id, value)}
        >
          <Option value="todo">待开始</Option>
          <Option value="in_progress">进行中</Option>
          <Option value="review">待审核</Option>
          <Option value="completed">已完成</Option>
        </Select>
      )
    },
    {
      title: '优先级',
      dataIndex: 'priority',
      key: 'priority',
      render: (priority) => (
        <Tag color={getPriorityColor(priority)}>
          <FlagOutlined /> {getPriorityText(priority)}
        </Tag>
      )
    },
    {
      title: '负责人',
      dataIndex: 'assignee',
      key: 'assignee',
      render: (assignee) => (
        <Space>
          <Avatar size="small" icon={<UserOutlined />}>
            {assignee.name.charAt(0)}
          </Avatar>
          <span>{assignee.name}</span>
        </Space>
      )
    },
    {
      title: '所属项目',
      dataIndex: 'project',
      key: 'project',
      render: (project) => project.name
    },
    {
      title: '截止时间',
      dataIndex: 'dueDate',
      key: 'dueDate',
      render: (date) => {
        const isOverdue = dayjs(date).isBefore(dayjs(), 'day');
        return (
          <span style={{ color: isOverdue ? '#ff4d4f' : undefined }}>
            <CalendarOutlined /> {dayjs(date).format('MM-DD')}
          </span>
        );
      }
    },
    {
      title: '进度',
      key: 'progress',
      render: (_, record) => {
        if (record.subtasks) {
          const completed = record.subtasks.filter(st => st.completed).length;
          const total = record.subtasks.length;
          const percent = Math.round((completed / total) * 100);
          return <Progress percent={percent} size="small" />;
        }
        return record.status === 'completed' ? 
          <Progress percent={100} size="small" status="success" /> :
          <Progress percent={0} size="small" />;
      }
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => (
        <Dropdown
          overlay={
            <Menu onClick={({ key }) => handleMenuClick(key, record)}>
              <Menu.Item key="view" icon={<EyeOutlined />}>
                查看详情
              </Menu.Item>
              <Menu.Item key="edit" icon={<EditOutlined />}>
                编辑任务
              </Menu.Item>
              <Menu.Item key="delete" icon={<DeleteOutlined />} danger>
                删除任务
              </Menu.Item>
            </Menu>
          }
          trigger={['click']}
        >
          <Button type="text" icon={<MoreOutlined />} />
        </Dropdown>
      )
    }
  ];

  const getFilteredTasks = () => {
    switch (activeTab) {
      case 'todo':
        return tasks.filter(task => task.status === 'todo');
      case 'in_progress':
        return tasks.filter(task => task.status === 'in_progress');
      case 'review':
        return tasks.filter(task => task.status === 'review');
      case 'completed':
        return tasks.filter(task => task.status === 'completed');
      case 'my_tasks':
        return tasks.filter(task => task.assignee.name === '当前用户');
      default:
        return tasks;
    }
  };

  // 统计数据
  const stats = {
    total: tasks.length,
    todo: tasks.filter(t => t.status === 'todo').length,
    inProgress: tasks.filter(t => t.status === 'in_progress').length,
    review: tasks.filter(t => t.status === 'review').length,
    completed: tasks.filter(t => t.status === 'completed').length,
    overdue: tasks.filter(t => dayjs(t.dueDate).isBefore(dayjs(), 'day') && t.status !== 'completed').length
  };

  return (
    <div style={{ padding: '24px' }}>
      {/* 页面标题和操作 */}
      <div style={{ marginBottom: '24px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <h1 style={{ margin: 0, fontSize: '24px', fontWeight: 600 }}>任务管理</h1>
          <p style={{ margin: '8px 0 0 0', color: '#666' }}>
            创建、分配和跟踪项目任务，提升团队协作效率
          </p>
        </div>
        <Space>
          <Button icon={<FilterOutlined />}>筛选</Button>
          <Button 
            type="primary" 
            icon={<PlusOutlined />}
            onClick={() => setCreateModalVisible(true)}
          >
            创建任务
          </Button>
        </Space>
      </div>

      {/* 统计卡片 */}
      <Row gutter={16} style={{ marginBottom: '24px' }}>
        <Col span={4}>
          <Card size="small">
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: '24px', fontWeight: 600, color: '#1890ff' }}>
                {stats.total}
              </div>
              <div style={{ color: '#666' }}>总任务</div>
            </div>
          </Card>
        </Col>
        <Col span={4}>
          <Card size="small">
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: '24px', fontWeight: 600, color: '#faad14' }}>
                {stats.todo}
              </div>
              <div style={{ color: '#666' }}>待开始</div>
            </div>
          </Card>
        </Col>
        <Col span={4}>
          <Card size="small">
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: '24px', fontWeight: 600, color: '#1890ff' }}>
                {stats.inProgress}
              </div>
              <div style={{ color: '#666' }}>进行中</div>
            </div>
          </Card>
        </Col>
        <Col span={4}>
          <Card size="small">
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: '24px', fontWeight: 600, color: '#fa8c16' }}>
                {stats.review}
              </div>
              <div style={{ color: '#666' }}>待审核</div>
            </div>
          </Card>
        </Col>
        <Col span={4}>
          <Card size="small">
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: '24px', fontWeight: 600, color: '#52c41a' }}>
                {stats.completed}
              </div>
              <div style={{ color: '#666' }}>已完成</div>
            </div>
          </Card>
        </Col>
        <Col span={4}>
          <Card size="small">
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: '24px', fontWeight: 600, color: '#ff4d4f' }}>
                {stats.overdue}
              </div>
              <div style={{ color: '#666' }}>已逾期</div>
            </div>
          </Card>
        </Col>
      </Row>

      {/* 任务列表 */}
      <Card>
        <Tabs activeKey={activeTab} onChange={setActiveTab}>
          <TabPane tab={`全部任务 (${stats.total})`} key="all" />
          <TabPane tab={`待开始 (${stats.todo})`} key="todo" />
          <TabPane tab={`进行中 (${stats.inProgress})`} key="in_progress" />
          <TabPane tab={`待审核 (${stats.review})`} key="review" />
          <TabPane tab={`已完成 (${stats.completed})`} key="completed" />
          <TabPane tab="我的任务" key="my_tasks" />
        </Tabs>

        <Table
          columns={columns}
          dataSource={getFilteredTasks()}
          rowKey="id"
          loading={loading}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 个任务`
          }}
        />
      </Card>

      {/* 创建任务模态框 */}
      <Modal
        title="创建新任务"
        open={createModalVisible}
        onCancel={() => {
          setCreateModalVisible(false);
          form.resetFields();
        }}
        footer={null}
        width={600}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleCreateTask}
        >
          <Form.Item
            name="title"
            label="任务标题"
            rules={[{ required: true, message: '请输入任务标题' }]}
          >
            <Input placeholder="请输入任务标题" />
          </Form.Item>

          <Form.Item
            name="description"
            label="任务描述"
            rules={[{ required: true, message: '请输入任务描述' }]}
          >
            <TextArea rows={3} placeholder="请输入任务描述" />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="priority"
                label="优先级"
                rules={[{ required: true, message: '请选择优先级' }]}
              >
                <Select placeholder="请选择优先级">
                  <Option value="low">低</Option>
                  <Option value="medium">中</Option>
                  <Option value="high">高</Option>
                  <Option value="urgent">紧急</Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="project"
                label="所属项目"
                rules={[{ required: true, message: '请选择项目' }]}
              >
                <Select placeholder="请选择项目">
                  <Option value="1">太上老君文化研究项目</Option>
                  <Option value="2">道德经现代解读</Option>
                  <Option value="3">智慧学习平台开发</Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="dueDate"
                label="截止时间"
                rules={[{ required: true, message: '请选择截止时间' }]}
              >
                <DatePicker style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="estimatedHours"
                label="预估工时(小时)"
                rules={[{ required: true, message: '请输入预估工时' }]}
              >
                <Input type="number" placeholder="请输入预估工时" />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            name="tags"
            label="标签"
          >
            <Select mode="tags" placeholder="请输入标签">
              <Option value="文献">文献</Option>
              <Option value="研究">研究</Option>
              <Option value="开发">开发</Option>
              <Option value="设计">设计</Option>
              <Option value="测试">测试</Option>
            </Select>
          </Form.Item>

          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={() => {
                setCreateModalVisible(false);
                form.resetFields();
              }}>
                取消
              </Button>
              <Button type="primary" htmlType="submit" loading={loading}>
                创建任务
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default TaskManagement;