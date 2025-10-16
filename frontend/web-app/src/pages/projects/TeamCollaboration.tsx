import React, { useState, useEffect } from 'react';
import {
  Card,
  Row,
  Col,
  Button,
  Table,
  Tag,
  Space,
  Avatar,
  Tooltip,
  Modal,
  Form,
  Input,
  Select,
  message,
  Dropdown,
  Menu,
  Tabs,
  List,
  Upload,
  Progress,
  Timeline,
  Badge,
  Divider
} from 'antd';
import {
  PlusOutlined,
  TeamOutlined,
  UserOutlined,
  FileTextOutlined,
  ShareAltOutlined,
  MoreOutlined,
  EditOutlined,
  DeleteOutlined,
  DownloadOutlined,
  UploadOutlined,
  MessageOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';

const { Option } = Select;
const { TextArea } = Input;


interface TeamMember {
  id: string;
  name: string;
  email: string;
  role: 'owner' | 'admin' | 'member' | 'viewer';
  avatar?: string;
  status: 'online' | 'offline' | 'busy';
  joinDate: string;
  lastActive: string;
  projects: string[];
  skills: string[];
}

interface Document {
  id: string;
  name: string;
  type: 'document' | 'spreadsheet' | 'presentation' | 'pdf' | 'image';
  size: number;
  author: string;
  lastModified: string;
  sharedWith: string[];
  version: number;
  status: 'draft' | 'review' | 'approved';
}

interface Activity {
  id: string;
  type: 'document_created' | 'document_updated' | 'member_joined' | 'task_completed' | 'comment_added';
  user: string;
  description: string;
  timestamp: string;
  target?: string;
}

const TeamCollaboration: React.FC = () => {
  const [members, setMembers] = useState<TeamMember[]>([]);
  const [documents, setDocuments] = useState<Document[]>([]);
  const [activities, setActivities] = useState<Activity[]>([]);
  const [loading, setLoading] = useState(false);
  const [inviteModalVisible, setInviteModalVisible] = useState(false);
  const [activeTab, setActiveTab] = useState('members');
  const [form] = Form.useForm();

  // 模拟数据
  useEffect(() => {
    setLoading(true);
    setTimeout(() => {
      setMembers([
        {
          id: '1',
          name: '张三',
          email: 'zhangsan@example.com',
          role: 'owner',
          status: 'online',
          joinDate: '2024-01-01',
          lastActive: '2024-02-01 10:30',
          projects: ['太上老君文化研究项目', '智慧学习平台开发'],
          skills: ['项目管理', '文献研究', '团队协调']
        },
        {
          id: '2',
          name: '李四',
          email: 'lisi@example.com',
          role: 'admin',
          status: 'online',
          joinDate: '2024-01-05',
          lastActive: '2024-02-01 09:45',
          projects: ['太上老君文化研究项目'],
          skills: ['学术研究', '文献整理', '写作']
        },
        {
          id: '3',
          name: '王五',
          email: 'wangwu@example.com',
          role: 'member',
          status: 'busy',
          joinDate: '2024-01-10',
          lastActive: '2024-02-01 08:20',
          projects: ['智慧学习平台开发'],
          skills: ['UI设计', '用户体验', '原型设计']
        },
        {
          id: '4',
          name: '赵六',
          email: 'zhaoliu@example.com',
          role: 'member',
          status: 'offline',
          joinDate: '2024-01-15',
          lastActive: '2024-01-31 18:00',
          projects: ['智慧学习平台开发'],
          skills: ['前端开发', 'React', 'TypeScript']
        }
      ]);

      setDocuments([
        {
          id: '1',
          name: '太上老君思想体系研究报告.docx',
          type: 'document',
          size: 2048000,
          author: '李四',
          lastModified: '2024-02-01 14:30',
          sharedWith: ['张三', '王五'],
          version: 3,
          status: 'review'
        },
        {
          id: '2',
          name: '项目进度跟踪表.xlsx',
          type: 'spreadsheet',
          size: 512000,
          author: '张三',
          lastModified: '2024-02-01 11:20',
          sharedWith: ['李四', '王五', '赵六'],
          version: 5,
          status: 'approved'
        },
        {
          id: '3',
          name: '学习平台UI设计稿.pptx',
          type: 'presentation',
          size: 8192000,
          author: '王五',
          lastModified: '2024-01-31 16:45',
          sharedWith: ['张三', '赵六'],
          version: 2,
          status: 'draft'
        }
      ]);

      setActivities([
        {
          id: '1',
          type: 'document_updated',
          user: '李四',
          description: '更新了文档《太上老君思想体系研究报告》',
          timestamp: '2024-02-01 14:30',
          target: '太上老君思想体系研究报告.docx'
        },
        {
          id: '2',
          type: 'task_completed',
          user: '王五',
          description: '完成了任务《设计学习平台UI界面》',
          timestamp: '2024-02-01 12:15',
          target: '设计学习平台UI界面'
        },
        {
          id: '3',
          type: 'comment_added',
          user: '张三',
          description: '在项目《太上老君文化研究项目》中添加了评论',
          timestamp: '2024-02-01 10:30',
          target: '太上老君文化研究项目'
        },
        {
          id: '4',
          type: 'document_created',
          user: '赵六',
          description: '创建了新文档《技术架构设计》',
          timestamp: '2024-01-31 18:00',
          target: '技术架构设计.pdf'
        }
      ]);

      setLoading(false);
    }, 1000);
  }, []);

  const getRoleColor = (role: TeamMember['role']) => {
    const colors = {
      owner: 'red',
      admin: 'orange',
      member: 'blue',
      viewer: 'default'
    };
    return colors[role];
  };

  const getRoleText = (role: TeamMember['role']) => {
    const texts = {
      owner: '所有者',
      admin: '管理员',
      member: '成员',
      viewer: '查看者'
    };
    return texts[role];
  };

  const getStatusColor = (status: TeamMember['status']) => {
    const colors = {
      online: 'green',
      offline: 'default',
      busy: 'orange'
    };
    return colors[status];
  };

  const getStatusText = (status: TeamMember['status']) => {
    const texts = {
      online: '在线',
      offline: '离线',
      busy: '忙碌'
    };
    return texts[status];
  };

  const getDocumentIcon = (type: Document['type']) => {
    const icons = {
      document: <FileTextOutlined style={{ color: '#1890ff' }} />,
      spreadsheet: <FileTextOutlined style={{ color: '#52c41a' }} />,
      presentation: <FileTextOutlined style={{ color: '#fa8c16' }} />,
      pdf: <FileTextOutlined style={{ color: '#f5222d' }} />,
      image: <FileTextOutlined style={{ color: '#722ed1' }} />
    };
    return icons[type];
  };

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const handleInviteMember = async (values: any) => {
    try {
      setLoading(true);
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      const newMember: TeamMember = {
        id: Date.now().toString(),
        name: values.name,
        email: values.email,
        role: values.role,
        status: 'offline',
        joinDate: dayjs().format('YYYY-MM-DD'),
        lastActive: dayjs().format('YYYY-MM-DD HH:mm'),
        projects: [],
        skills: values.skills || []
      };

      setMembers([...members, newMember]);
      setInviteModalVisible(false);
      form.resetFields();
      message.success('团队成员邀请成功！');
    } catch (error) {
      message.error('邀请失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  const handleRoleChange = (memberId: string, newRole: TeamMember['role']) => {
    setMembers(members.map(member => 
      member.id === memberId ? { ...member, role: newRole } : member
    ));
    message.success('角色更新成功');
  };

  const memberColumns: ColumnsType<TeamMember> = [
    {
      title: '成员',
      key: 'member',
      render: (_, record) => (
        <Space>
          <Badge 
            status={getStatusColor(record.status) as any} 
            dot
          >
            <Avatar icon={<UserOutlined />}>
              {record.name.charAt(0)}
            </Avatar>
          </Badge>
          <div>
            <div style={{ fontWeight: 500 }}>{record.name}</div>
            <div style={{ fontSize: '12px', color: '#666' }}>{record.email}</div>
          </div>
        </Space>
      )
    },
    {
      title: '角色',
      dataIndex: 'role',
      key: 'role',
      render: (role, record) => (
        <Select
          value={role}
          size="small"
          style={{ width: 100 }}
          onChange={(value) => handleRoleChange(record.id, value)}
        >
          <Option value="owner">所有者</Option>
          <Option value="admin">管理员</Option>
          <Option value="member">成员</Option>
          <Option value="viewer">查看者</Option>
        </Select>
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
      title: '技能',
      dataIndex: 'skills',
      key: 'skills',
      render: (skills) => (
        <Space size={4} wrap>
          {skills.slice(0, 2).map((skill: string) => (
            <Tag key={skill} size="small">{skill}</Tag>
          ))}
          {skills.length > 2 && (
            <Tag size="small">+{skills.length - 2}</Tag>
          )}
        </Space>
      )
    },
    {
      title: '最后活跃',
      dataIndex: 'lastActive',
      key: 'lastActive',
      render: (time) => dayjs(time).format('MM-DD HH:mm')
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => (
        <Dropdown
          overlay={
            <Menu>
              <Menu.Item key="message" icon={<MessageOutlined />}>
                发送消息
              </Menu.Item>
              <Menu.Item key="projects" icon={<TeamOutlined />}>
                查看项目
              </Menu.Item>
              <Menu.Item key="remove" icon={<DeleteOutlined />} danger>
                移除成员
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

  const documentColumns: ColumnsType<Document> = [
    {
      title: '文档名称',
      key: 'name',
      render: (_, record) => (
        <Space>
          {getDocumentIcon(record.type)}
          <div>
            <div style={{ fontWeight: 500 }}>{record.name}</div>
            <div style={{ fontSize: '12px', color: '#666' }}>
              v{record.version} · {formatFileSize(record.size)}
            </div>
          </div>
        </Space>
      )
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status) => {
        const statusConfig = {
          draft: { color: 'default', text: '草稿' },
          review: { color: 'processing', text: '审核中' },
          approved: { color: 'success', text: '已批准' }
        };
        const config = statusConfig[status];
        return <Tag color={config.color}>{config.text}</Tag>;
      }
    },
    {
      title: '作者',
      dataIndex: 'author',
      key: 'author'
    },
    {
      title: '共享对象',
      dataIndex: 'sharedWith',
      key: 'sharedWith',
      render: (sharedWith) => (
        <Avatar.Group maxCount={3} size="small">
          {sharedWith.map((name: string, index: number) => (
            <Tooltip key={index} title={name}>
              <Avatar size="small">{name.charAt(0)}</Avatar>
            </Tooltip>
          ))}
        </Avatar.Group>
      )
    },
    {
      title: '最后修改',
      dataIndex: 'lastModified',
      key: 'lastModified',
      render: (time) => dayjs(time).format('MM-DD HH:mm')
    },
    {
      title: '操作',
      key: 'action',
      render: () => (
        <Space>
          <Button type="text" icon={<DownloadOutlined />} />
          <Button type="text" icon={<ShareAltOutlined />} />
          <Button type="text" icon={<MoreOutlined />} />
        </Space>
      )
    }
  ];

  const getActivityIcon = (type: Activity['type']) => {
    const icons = {
      document_created: <FileTextOutlined style={{ color: '#52c41a' }} />,
      document_updated: <EditOutlined style={{ color: '#1890ff' }} />,
      member_joined: <UserOutlined style={{ color: '#722ed1' }} />,
      task_completed: <CheckCircleOutlined style={{ color: '#52c41a' }} />,
      comment_added: <MessageOutlined style={{ color: '#fa8c16' }} />
    };
    return icons[type];
  };

  return (
    <div style={{ padding: '24px' }}>
      {/* 页面标题和操作 */}
      <div style={{ marginBottom: '24px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <h1 style={{ margin: 0, fontSize: '24px', fontWeight: 600 }}>团队协作</h1>
          <p style={{ margin: '8px 0 0 0', color: '#666' }}>
            管理团队成员，共享文档资源，实时协作沟通
          </p>
        </div>
        <Space>
          <Button icon={<UploadOutlined />}>上传文档</Button>
          <Button 
            type="primary" 
            icon={<PlusOutlined />}
            onClick={() => setInviteModalVisible(true)}
          >
            邀请成员
          </Button>
        </Space>
      </div>

      {/* 统计卡片 */}
      <Row gutter={16} style={{ marginBottom: '24px' }}>
        <Col span={6}>
          <Card size="small">
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: '24px', fontWeight: 600, color: '#1890ff' }}>
                {members.length}
              </div>
              <div style={{ color: '#666' }}>团队成员</div>
            </div>
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small">
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: '24px', fontWeight: 600, color: '#52c41a' }}>
                {members.filter(m => m.status === 'online').length}
              </div>
              <div style={{ color: '#666' }}>在线成员</div>
            </div>
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small">
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: '24px', fontWeight: 600, color: '#722ed1' }}>
                {documents.length}
              </div>
              <div style={{ color: '#666' }}>共享文档</div>
            </div>
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small">
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: '24px', fontWeight: 600, color: '#fa8c16' }}>
                {activities.length}
              </div>
              <div style={{ color: '#666' }}>今日活动</div>
            </div>
          </Card>
        </Col>
      </Row>

      {/* 主要内容区域 */}
      <Row gutter={16}>
        <Col span={16}>
          <Card>
            <Tabs 
              activeKey={activeTab} 
              onChange={setActiveTab}
              items={[
                {
                  key: 'members',
                  label: `团队成员 (${members.length})`,
                  children: (
                    <Table
                      columns={memberColumns}
                      dataSource={members}
                      rowKey="id"
                      loading={loading}
                      pagination={false}
                    />
                  )
                },
                {
                  key: 'documents',
                  label: `共享文档 (${documents.length})`,
                  children: (
                    <Table
                      columns={documentColumns}
                      dataSource={documents}
                      rowKey="id"
                      loading={loading}
                      pagination={false}
                    />
                  )
                }
              ]}
            />
          </Card>
        </Col>
        
        <Col span={8}>
          <Card title="团队动态" extra={<Button type="link">查看全部</Button>}>
            <Timeline>
              {activities.map(activity => (
                <Timeline.Item
                  key={activity.id}
                  dot={getActivityIcon(activity.type)}
                >
                  <div style={{ fontSize: '14px' }}>
                    <strong>{activity.user}</strong> {activity.description}
                  </div>
                  <div style={{ fontSize: '12px', color: '#666', marginTop: '4px' }}>
                    {dayjs(activity.timestamp).format('MM-DD HH:mm')}
                  </div>
                </Timeline.Item>
              ))}
            </Timeline>
          </Card>
        </Col>
      </Row>

      {/* 邀请成员模态框 */}
      <Modal
        title="邀请团队成员"
        open={inviteModalVisible}
        onCancel={() => {
          setInviteModalVisible(false);
          form.resetFields();
        }}
        footer={null}
        width={500}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleInviteMember}
        >
          <Form.Item
            name="name"
            label="姓名"
            rules={[{ required: true, message: '请输入姓名' }]}
          >
            <Input placeholder="请输入姓名" />
          </Form.Item>

          <Form.Item
            name="email"
            label="邮箱"
            rules={[
              { required: true, message: '请输入邮箱' },
              { type: 'email', message: '请输入有效的邮箱地址' }
            ]}
          >
            <Input placeholder="请输入邮箱地址" />
          </Form.Item>

          <Form.Item
            name="role"
            label="角色"
            rules={[{ required: true, message: '请选择角色' }]}
          >
            <Select placeholder="请选择角色">
              <Option value="member">成员</Option>
              <Option value="admin">管理员</Option>
              <Option value="viewer">查看者</Option>
            </Select>
          </Form.Item>

          <Form.Item
            name="skills"
            label="技能标签"
          >
            <Select mode="tags" placeholder="请输入技能标签">
              <Option value="项目管理">项目管理</Option>
              <Option value="文献研究">文献研究</Option>
              <Option value="前端开发">前端开发</Option>
              <Option value="后端开发">后端开发</Option>
              <Option value="UI设计">UI设计</Option>
              <Option value="数据分析">数据分析</Option>
            </Select>
          </Form.Item>

          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={() => {
                setInviteModalVisible(false);
                form.resetFields();
              }}>
                取消
              </Button>
              <Button type="primary" htmlType="submit" loading={loading}>
                发送邀请
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default TeamCollaboration;