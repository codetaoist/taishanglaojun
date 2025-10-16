import React, { useState, useEffect } from 'react';
import {
  Card,
  Row,
  Col,
  Button,
  Table,
  Tag,
  Progress,
  Statistic,
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
  Menu
} from 'antd';
import {
  PlusOutlined,
  ProjectOutlined,
  TeamOutlined,
  CalendarOutlined,
  BarChartOutlined,
  MoreOutlined,
  EditOutlined,
  DeleteOutlined,
  EyeOutlined,
  UserOutlined
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';

const { Option } = Select;
const { RangePicker } = DatePicker;
const { TextArea } = Input;

interface Project {
  id: string;
  name: string;
  description: string;
  status: 'planning' | 'active' | 'completed' | 'paused';
  priority: 'low' | 'medium' | 'high';
  progress: number;
  startDate: string;
  endDate: string;
  teamMembers: Array<{
    id: string;
    name: string;
    avatar?: string;
    role: string;
  }>;
  tasksCount: number;
  completedTasks: number;
  createdBy: string;
  createdAt: string;
}

const ProjectWorkspace: React.FC = () => {
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(false);
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [form] = Form.useForm();

  // 模拟数据
  useEffect(() => {
    setLoading(true);
    setTimeout(() => {
      setProjects([
        {
          id: '1',
          name: '太上老君文化研究项目',
          description: '深入研究太上老君文化内涵，整理相关文献资料',
          status: 'active',
          priority: 'high',
          progress: 65,
          startDate: '2024-01-15',
          endDate: '2024-06-30',
          teamMembers: [
            { id: '1', name: '张三', role: '项目经理' },
            { id: '2', name: '李四', role: '研究员' },
            { id: '3', name: '王五', role: '文档整理' }
          ],
          tasksCount: 24,
          completedTasks: 16,
          createdBy: '张三',
          createdAt: '2024-01-15'
        },
        {
          id: '2',
          name: '道德经现代解读',
          description: '结合现代社会背景，重新解读道德经的智慧',
          status: 'planning',
          priority: 'medium',
          progress: 20,
          startDate: '2024-03-01',
          endDate: '2024-08-31',
          teamMembers: [
            { id: '4', name: '赵六', role: '项目经理' },
            { id: '5', name: '钱七', role: '研究员' }
          ],
          tasksCount: 18,
          completedTasks: 3,
          createdBy: '赵六',
          createdAt: '2024-02-20'
        },
        {
          id: '3',
          name: '智慧学习平台开发',
          description: '开发基于AI的智慧学习平台，提升学习效率',
          status: 'completed',
          priority: 'high',
          progress: 100,
          startDate: '2023-10-01',
          endDate: '2024-01-31',
          teamMembers: [
            { id: '6', name: '孙八', role: '技术负责人' },
            { id: '7', name: '周九', role: '前端开发' },
            { id: '8', name: '吴十', role: '后端开发' }
          ],
          tasksCount: 45,
          completedTasks: 45,
          createdBy: '孙八',
          createdAt: '2023-09-15'
        }
      ]);
      setLoading(false);
    }, 1000);
  }, []);

  const getStatusColor = (status: Project['status']) => {
    const colors = {
      planning: 'blue',
      active: 'green',
      completed: 'gray',
      paused: 'orange'
    };
    return colors[status];
  };

  const getStatusText = (status: Project['status']) => {
    const texts = {
      planning: '规划中',
      active: '进行中',
      completed: '已完成',
      paused: '已暂停'
    };
    return texts[status];
  };

  const getPriorityColor = (priority: Project['priority']) => {
    const colors = {
      low: 'green',
      medium: 'orange',
      high: 'red'
    };
    return colors[priority];
  };

  const getPriorityText = (priority: Project['priority']) => {
    const texts = {
      low: '低',
      medium: '中',
      high: '高'
    };
    return texts[priority];
  };

  const handleCreateProject = async (values: any) => {
    try {
      setLoading(true);
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      const newProject: Project = {
        id: Date.now().toString(),
        name: values.name,
        description: values.description,
        status: 'planning',
        priority: values.priority,
        progress: 0,
        startDate: values.dateRange[0].format('YYYY-MM-DD'),
        endDate: values.dateRange[1].format('YYYY-MM-DD'),
        teamMembers: [],
        tasksCount: 0,
        completedTasks: 0,
        createdBy: '当前用户',
        createdAt: dayjs().format('YYYY-MM-DD')
      };

      setProjects([newProject, ...projects]);
      setCreateModalVisible(false);
      form.resetFields();
      message.success('项目创建成功！');
    } catch (error) {
      message.error('项目创建失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  const handleMenuClick = (key: string, project: Project) => {
    switch (key) {
      case 'view':
        message.info(`查看项目: ${project.name}`);
        break;
      case 'edit':
        message.info(`编辑项目: ${project.name}`);
        break;
      case 'delete':
        Modal.confirm({
          title: '确认删除',
          content: `确定要删除项目"${project.name}"吗？`,
          onOk: () => {
            setProjects(projects.filter(p => p.id !== project.id));
            message.success('项目删除成功');
          }
        });
        break;
    }
  };

  const columns: ColumnsType<Project> = [
    {
      title: '项目名称',
      dataIndex: 'name',
      key: 'name',
      render: (text, record) => (
        <Space>
          <ProjectOutlined />
          <span style={{ fontWeight: 500 }}>{text}</span>
        </Space>
      )
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status) => (
        <Tag color={getStatusColor(status)}>
          {getStatusText(status)}
        </Tag>
      )
    },
    {
      title: '优先级',
      dataIndex: 'priority',
      key: 'priority',
      render: (priority) => (
        <Tag color={getPriorityColor(priority)}>
          {getPriorityText(priority)}
        </Tag>
      )
    },
    {
      title: '进度',
      dataIndex: 'progress',
      key: 'progress',
      render: (progress) => (
        <Progress 
          percent={progress} 
          size="small" 
          status={progress === 100 ? 'success' : 'active'}
        />
      )
    },
    {
      title: '团队成员',
      dataIndex: 'teamMembers',
      key: 'teamMembers',
      render: (members) => (
        <Avatar.Group maxCount={3} size="small">
          {members.map((member: any) => (
            <Tooltip key={member.id} title={`${member.name} - ${member.role}`}>
              <Avatar icon={<UserOutlined />}>
                {member.name.charAt(0)}
              </Avatar>
            </Tooltip>
          ))}
        </Avatar.Group>
      )
    },
    {
      title: '任务进度',
      key: 'tasks',
      render: (_, record) => (
        <span>{record.completedTasks}/{record.tasksCount}</span>
      )
    },
    {
      title: '结束时间',
      dataIndex: 'endDate',
      key: 'endDate',
      render: (date) => dayjs(date).format('YYYY-MM-DD')
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
                编辑项目
              </Menu.Item>
              <Menu.Item key="delete" icon={<DeleteOutlined />} danger>
                删除项目
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

  // 统计数据
  const stats = {
    total: projects.length,
    active: projects.filter(p => p.status === 'active').length,
    completed: projects.filter(p => p.status === 'completed').length,
    avgProgress: projects.length > 0 
      ? Math.round(projects.reduce((sum, p) => sum + p.progress, 0) / projects.length)
      : 0
  };

  return (
    <div style={{ padding: '24px' }}>
      {/* 页面标题和操作 */}
      <div style={{ marginBottom: '24px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <h1 style={{ margin: 0, fontSize: '24px', fontWeight: 600 }}>项目工作台</h1>
          <p style={{ margin: '8px 0 0 0', color: '#666' }}>
            管理您的所有项目，跟踪进度，协调团队协作
          </p>
        </div>
        <Button 
          type="primary" 
          icon={<PlusOutlined />}
          onClick={() => setCreateModalVisible(true)}
        >
          创建项目
        </Button>
      </div>

      {/* 统计卡片 */}
      <Row gutter={16} style={{ marginBottom: '24px' }}>
        <Col span={6}>
          <Card>
            <Statistic
              title="总项目数"
              value={stats.total}
              prefix={<ProjectOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="进行中"
              value={stats.active}
              prefix={<CalendarOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="已完成"
              value={stats.completed}
              prefix={<BarChartOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="平均进度"
              value={stats.avgProgress}
              suffix="%"
              prefix={<TeamOutlined />}
              valueStyle={{ color: '#fa8c16' }}
            />
          </Card>
        </Col>
      </Row>

      {/* 项目列表 */}
      <Card title="项目列表" extra={<Button type="link">查看全部</Button>}>
        <Table
          columns={columns}
          dataSource={projects}
          rowKey="id"
          loading={loading}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 个项目`
          }}
        />
      </Card>

      {/* 创建项目模态框 */}
      <Modal
        title="创建新项目"
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
          onFinish={handleCreateProject}
        >
          <Form.Item
            name="name"
            label="项目名称"
            rules={[{ required: true, message: '请输入项目名称' }]}
          >
            <Input placeholder="请输入项目名称" />
          </Form.Item>

          <Form.Item
            name="description"
            label="项目描述"
            rules={[{ required: true, message: '请输入项目描述' }]}
          >
            <TextArea rows={3} placeholder="请输入项目描述" />
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
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="dateRange"
                label="项目周期"
                rules={[{ required: true, message: '请选择项目周期' }]}
              >
                <RangePicker style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={() => {
                setCreateModalVisible(false);
                form.resetFields();
              }}>
                取消
              </Button>
              <Button type="primary" htmlType="submit" loading={loading}>
                创建项目
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default ProjectWorkspace;