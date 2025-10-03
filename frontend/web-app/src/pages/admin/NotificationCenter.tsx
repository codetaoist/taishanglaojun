import React, { useState, useEffect } from 'react';
import {
  Card,
  Table,
  Button,
  Space,
  Tag,
  Modal,
  Form,
  Input,
  Select,
  DatePicker,
  message,
  Tabs,
  Avatar,
  Row,
  Col,
  Statistic,
  Progress,
  Tooltip,
  Popconfirm
} from 'antd';
import {
  BellOutlined,
  SendOutlined,
  DeleteOutlined,
  EditOutlined,
  EyeOutlined,
  UserOutlined,
  CheckCircleOutlined,
  NotificationOutlined,
  PlusOutlined,
  ReloadOutlined
} from '@ant-design/icons';
import type { ColumnType } from 'antd/es/table';
import { apiClient } from '../../services/api';
import dayjs from 'dayjs';

const { TabPane } = Tabs;
const { TextArea } = Input;
const { Option } = Select;

interface Notification {
  id: string;
  title: string;
  content: string;
  type: 'system' | 'user' | 'announcement' | 'warning';
  priority: 'low' | 'medium' | 'high' | 'urgent';
  status: 'draft' | 'sent' | 'scheduled';
  targetType: 'all' | 'role' | 'user' | 'group';
  targetValue?: string;
  sendTime?: string;
  readCount: number;
  totalCount: number;
  createdAt: string;
  createdBy: string;
}

interface NotificationStats {
  totalNotifications: number;
  sentNotifications: number;
  draftNotifications: number;
  scheduledNotifications: number;
  totalReads: number;
  averageReadRate: number;
}

interface UserNotification {
  id: string;
  userId: string;
  userName: string;
  userAvatar: string;
  notificationId: string;
  notificationTitle: string;
  isRead: boolean;
  readTime?: string;
  createdAt: string;
}

const NotificationCenter: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [userNotifications, setUserNotifications] = useState<UserNotification[]>([]);
  const [stats, setStats] = useState<NotificationStats>({
    totalNotifications: 0,
    sentNotifications: 0,
    draftNotifications: 0,
    scheduledNotifications: 0,
    totalReads: 0,
    averageReadRate: 0
  });
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [editingNotification, setEditingNotification] = useState<Notification | null>(null);
  const [form] = Form.useForm();
  const [activeTab, setActiveTab] = useState('notifications');

  useEffect(() => {
    loadNotifications();
    loadUserNotifications();
    loadStats();
  }, []);

  const loadNotifications = async () => {
    setLoading(true);
    try {
      const response = await apiClient.getNotifications();
      setNotifications(response.data.items || []);
    } catch {
      message.error('加载通知列表失败');
    } finally {
      setLoading(false);
    }
  };

  const loadUserNotifications = async () => {
    try {
      const response = await apiClient.getUserNotifications();
      setUserNotifications(response.data.items || []);
    } catch {
      message.error('加载通知失败');
    }
  };

  const loadStats = async () => {
    try {
      const response = await apiClient.getNotificationStats();
      setStats(response.data);
    } catch (error) {
      console.error('加载通知统计失败:', error);
    }
  };

  const handleCreate = () => {
    setEditingNotification(null);
    form.resetFields();
    setIsModalVisible(true);
  };

  const handleEdit = (record: Notification) => {
    setEditingNotification(record);
    form.setFieldsValue({
      ...record,
      sendTime: record.sendTime ? dayjs(record.sendTime) : null
    });
    setIsModalVisible(true);
  };

  const handleDelete = async (id: string) => {
    try {
      await apiClient.deleteNotification(id);
      message.success('删除成功');
      loadNotifications();
      loadStats();
    } catch {
      message.error('删除通知失败');
    }
  };

  const handleBatchDelete = async () => {
    try {
      await apiClient.batchDeleteNotifications(selectedRowKeys as string[]);
      message.success('批量删除成功');
      setSelectedRowKeys([]);
      loadNotifications();
      loadStats();
    } catch {
      message.error('批量删除失败');
    }
  };

  const handleSend = async (id: string) => {
    try {
      await apiClient.sendNotification(id);
      message.success('发送成功');
      loadNotifications();
      loadStats();
    } catch {
      message.error('发送通知失败');
    }
  };

  const handleSubmit = async (values: {
    title: string;
    content: string;
    type: string;
    priority: string;
    targetType: string;
    targetValue?: string;
    sendTime?: dayjs.Dayjs;
  }) => {
    setLoading(true);
    try {
      const data = {
        ...values,
        sendTime: values.sendTime ? values.sendTime.format('YYYY-MM-DD HH:mm:ss') : null
      };

      if (editingNotification) {
        await apiClient.updateNotification(editingNotification.id, data);
        message.success('更新成功');
      } else {
        await apiClient.createNotification(data);
        message.success('创建成功');
      }

      setIsModalVisible(false);
      loadNotifications();
      loadStats();
    } catch {
      message.error(editingNotification ? '更新失败' : '创建失败');
    } finally {
      setLoading(false);
    }
  };

  const getTypeColor = (type: string) => {
    const colors = {
      system: 'blue',
      user: 'green',
      announcement: 'orange',
      warning: 'red'
    };
    return colors[type as keyof typeof colors] || 'default';
  };

  const getPriorityColor = (priority: string) => {
    const colors = {
      low: 'default',
      medium: 'blue',
      high: 'orange',
      urgent: 'red'
    };
    return colors[priority as keyof typeof colors] || 'default';
  };

  const getStatusColor = (status: string) => {
    const colors = {
      draft: 'default',
      sent: 'green',
      scheduled: 'blue'
    };
    return colors[status as keyof typeof colors] || 'default';
  };

  const notificationColumns: ColumnType<Notification>[] = [
    {
      title: '通知信息',
      dataIndex: 'title',
      key: 'title',
      render: (text, record) => (
        <div>
          <div style={{ fontWeight: 'bold', marginBottom: '4px' }}>{text}</div>
          <div style={{ color: '#666', fontSize: '12px' }}>
            {record.content.length > 50 ? `${record.content.substring(0, 50)}...` : record.content}
          </div>
        </div>
      )
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      width: 100,
      render: (type) => (
        <Tag color={getTypeColor(type)}>
          {type === 'system' && '系统'}
          {type === 'user' && '用户'}
          {type === 'announcement' && '公告'}
          {type === 'warning' && '警告'}
        </Tag>
      )
    },
    {
      title: '优先级',
      dataIndex: 'priority',
      key: 'priority',
      width: 100,
      render: (priority) => (
        <Tag color={getPriorityColor(priority)}>
          {priority === 'low' && '低'}
          {priority === 'medium' && '中'}
          {priority === 'high' && '高'}
          {priority === 'urgent' && '紧急'}
        </Tag>
      )
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status) => (
        <Tag color={getStatusColor(status)}>
          {status === 'draft' && '草稿'}
          {status === 'sent' && '已发送'}
          {status === 'scheduled' && '定时'}
        </Tag>
      )
    },
    {
      title: '阅读情况',
      key: 'readRate',
      width: 120,
      render: (_, record) => {
        const readRate = record.totalCount > 0 ? (record.readCount / record.totalCount * 100) : 0;
        return (
          <div>
            <div style={{ fontSize: '12px', color: '#666' }}>
              {record.readCount}/{record.totalCount}
            </div>
            <Progress percent={readRate} size="small" showInfo={false} />
          </div>
        );
      }
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      width: 150,
      render: (time) => dayjs(time).format('YYYY-MM-DD HH:mm')
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      render: (_, record) => (
        <Space size="small">
          <Tooltip title="查看">
            <Button
              type="text"
              icon={<EyeOutlined />}
              onClick={() => handleEdit(record)}
            />
          </Tooltip>
          <Tooltip title="编辑">
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => handleEdit(record)}
              disabled={record.status === 'sent'}
            />
          </Tooltip>
          {record.status === 'draft' && (
            <Tooltip title="发送">
              <Button
                type="text"
                icon={<SendOutlined />}
                onClick={() => handleSend(record.id)}
              />
            </Tooltip>
          )}
          <Popconfirm
            title="确定要删除这条通知吗？"
            onConfirm={() => handleDelete(record.id)}
          >
            <Tooltip title="删除">
              <Button
                type="text"
                danger
                icon={<DeleteOutlined />}
              />
            </Tooltip>
          </Popconfirm>
        </Space>
      )
    }
  ];

  const userNotificationColumns: ColumnType<UserNotification>[] = [
    {
      title: '用户',
      key: 'user',
      render: (_, record) => (
        <div style={{ display: 'flex', alignItems: 'center' }}>
          <Avatar src={record.userAvatar} icon={<UserOutlined />} />
          <span style={{ marginLeft: '8px' }}>{record.userName}</span>
        </div>
      )
    },
    {
      title: '通知标题',
      dataIndex: 'notificationTitle',
      key: 'notificationTitle'
    },
    {
      title: '状态',
      dataIndex: 'isRead',
      key: 'isRead',
      render: (isRead) => (
        <Tag color={isRead ? 'green' : 'orange'}>
          {isRead ? '已读' : '未读'}
        </Tag>
      )
    },
    {
      title: '阅读时间',
      dataIndex: 'readTime',
      key: 'readTime',
      render: (time) => time ? dayjs(time).format('YYYY-MM-DD HH:mm') : '-'
    },
    {
      title: '发送时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      render: (time) => dayjs(time).format('YYYY-MM-DD HH:mm')
    }
  ];

  const rowSelection = {
    selectedRowKeys,
    onChange: setSelectedRowKeys
  };

  return (
    <div style={{ padding: '24px' }}>
      <div style={{ marginBottom: '24px' }}>
        <h1 style={{ fontSize: '24px', fontWeight: 'bold', margin: 0 }}>
          <BellOutlined style={{ marginRight: '8px' }} />
          通知中心
        </h1>
        <p style={{ color: '#666', margin: '8px 0 0 0' }}>
          管理系统通知、用户消息和公告发布
        </p>
      </div>

      {/* 统计概览 */}
      <Row gutter={16} style={{ marginBottom: '24px' }}>
        <Col span={6}>
          <Card>
            <Statistic
              title="总通知数"
              value={stats.totalNotifications}
              prefix={<NotificationOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="已发送"
              value={stats.sentNotifications}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="草稿"
              value={stats.draftNotifications}
              prefix={<EditOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="平均阅读率"
              value={stats.averageReadRate}
              precision={2}
              suffix="%"
              prefix={<EyeOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
      </Row>

      <Tabs activeKey={activeTab} onChange={setActiveTab}>
        <TabPane
          tab={
            <span>
              <NotificationOutlined />
              通知管理
            </span>
          }
          key="notifications"
        >
          <Card>
            <div style={{ marginBottom: '16px', display: 'flex', justifyContent: 'space-between' }}>
              <Space>
                <Button
                  type="primary"
                  icon={<PlusOutlined />}
                  onClick={handleCreate}
                >
                  新建通知
                </Button>
                {selectedRowKeys.length > 0 && (
                  <Popconfirm
                    title={`确定要删除选中的 ${selectedRowKeys.length} 条通知吗？`}
                    onConfirm={handleBatchDelete}
                  >
                    <Button danger icon={<DeleteOutlined />}>
                      批量删除
                    </Button>
                  </Popconfirm>
                )}
                <Button icon={<ReloadOutlined />} onClick={loadNotifications}>
                  刷新
                </Button>
              </Space>
            </div>

            <Table
              rowSelection={rowSelection}
              columns={notificationColumns}
              dataSource={notifications}
              rowKey="id"
              loading={loading}
              pagination={{
                total: notifications.length,
                pageSize: 10,
                showSizeChanger: true,
                showQuickJumper: true,
                showTotal: (total) => `共 ${total} 条记录`
              }}
            />
          </Card>
        </TabPane>

        <TabPane
          tab={
            <span>
              <UserOutlined />
              用户通知
            </span>
          }
          key="userNotifications"
        >
          <Card>
            <div style={{ marginBottom: '16px' }}>
              <Button icon={<ReloadOutlined />} onClick={loadUserNotifications}>
                刷新
              </Button>
            </div>

            <Table
              columns={userNotificationColumns}
              dataSource={userNotifications}
              rowKey="id"
              loading={loading}
              pagination={{
                total: userNotifications.length,
                pageSize: 10,
                showSizeChanger: true,
                showQuickJumper: true,
                showTotal: (total) => `共 ${total} 条记录`
              }}
            />
          </Card>
        </TabPane>
      </Tabs>

      {/* 新建/编辑通知模态框 */}
      <Modal
        title={editingNotification ? '编辑通知' : '新建通知'}
        open={isModalVisible}
        onCancel={() => setIsModalVisible(false)}
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
                label="通知标题"
                name="title"
                rules={[{ required: true, message: '请输入通知标题' }]}
              >
                <Input placeholder="请输入通知标题" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="通知类型"
                name="type"
                rules={[{ required: true, message: '请选择通知类型' }]}
              >
                <Select placeholder="请选择通知类型">
                  <Option value="system">系统通知</Option>
                  <Option value="user">用户消息</Option>
                  <Option value="announcement">公告</Option>
                  <Option value="warning">警告</Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="优先级"
                name="priority"
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
                label="发送对象"
                name="targetType"
                rules={[{ required: true, message: '请选择发送对象' }]}
              >
                <Select placeholder="请选择发送对象">
                  <Option value="all">所有用户</Option>
                  <Option value="role">按角色</Option>
                  <Option value="user">指定用户</Option>
                  <Option value="group">用户组</Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            label="通知内容"
            name="content"
            rules={[{ required: true, message: '请输入通知内容' }]}
          >
            <TextArea rows={4} placeholder="请输入通知内容" />
          </Form.Item>

          <Form.Item
            label="定时发送"
            name="sendTime"
            help="不选择则立即发送"
          >
            <DatePicker
              showTime
              format="YYYY-MM-DD HH:mm:ss"
              placeholder="选择发送时间"
              style={{ width: '100%' }}
            />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={loading}>
                {editingNotification ? '更新' : '创建'}
              </Button>
              <Button onClick={() => setIsModalVisible(false)}>
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default NotificationCenter;