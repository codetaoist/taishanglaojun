import React, { useState, useEffect } from 'react';
import {
  Card,
  Form,
  Input,
  Button,
  Avatar,
  Upload,
  message,
  Row,
  Col,
  Divider,
  Typography,
  Space,
  Tabs,
  List,
  Tag,
  Statistic,
  Progress,
  Alert,
} from 'antd';
import {
  UserOutlined,
  CameraOutlined,
  EditOutlined,
  SaveOutlined,
  HistoryOutlined,
  TrophyOutlined,
  BookOutlined,
  MessageOutlined,
  UploadOutlined,
} from '@ant-design/icons';
import type { UploadProps, TabsProps } from 'antd';
import { useAuth } from '../hooks/useAuth';
import { apiClient } from '../services/api';
import dayjs from 'dayjs';

const { Title, Text, Paragraph } = Typography;
const { TextArea } = Input;

interface UserProfile {
  id: string;
  username: string;
  email: string;
  name?: string;
  avatar?: string;
  bio?: string;
  location?: string;
  website?: string;
  phone?: string;
  birthday?: string;
  gender?: 'male' | 'female' | 'other';
  createdAt: string;
  updatedAt: string;
  lastLogin?: string;
  loginCount: number;
  wisdomCount: number;
  commentCount: number;
  likeCount: number;
  followerCount: number;
  followingCount: number;
}

interface UserActivity {
  id: string;
  type: 'wisdom' | 'comment' | 'like' | 'follow';
  title: string;
  description: string;
  createdAt: string;
}

interface UserStats {
  totalWisdom: number;
  totalComments: number;
  totalLikes: number;
  totalViews: number;
  monthlyWisdom: number;
  monthlyComments: number;
}

const Profile: React.FC = () => {
  const { user, updateUser } = useAuth();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [editing, setEditing] = useState(false);
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [activities, setActivities] = useState<UserActivity[]>([]);
  const [stats, setStats] = useState<UserStats | null>(null);
  const [uploading, setUploading] = useState(false);

  // 获取用户资料
  const fetchProfile = async () => {
    try {
      setLoading(true);
      const response = await apiClient.getCurrentUser();
      setProfile(response.data);
      form.setFieldsValue(response.data);
    } catch (error) {
      message.error('获取用户资料失败');
    } finally {
      setLoading(false);
    }
  };

  // 获取用户活动
  const fetchActivities = async () => {
    try {
      // 模拟数据，实际应该调用API
      const mockActivities: UserActivity[] = [
        {
          id: '1',
          type: 'wisdom',
          title: '发布了新的智慧',
          description: '《道德经》第一章解读',
          createdAt: dayjs().subtract(1, 'hour').toISOString(),
        },
        {
          id: '2',
          type: 'comment',
          title: '评论了智慧',
          description: '对《论语》的深度思考',
          createdAt: dayjs().subtract(3, 'hours').toISOString(),
        },
        {
          id: '3',
          type: 'like',
          title: '点赞了智慧',
          description: '《庄子》逍遥游赏析',
          createdAt: dayjs().subtract(1, 'day').toISOString(),
        },
      ];
      setActivities(mockActivities);
    } catch (error) {
      message.error('获取用户活动失败');
    }
  };

  // 获取用户统计
  const fetchStats = async () => {
    try {
      // 模拟数据，实际应该调用API
      const mockStats: UserStats = {
        totalWisdom: 25,
        totalComments: 128,
        totalLikes: 456,
        totalViews: 1234,
        monthlyWisdom: 5,
        monthlyComments: 32,
      };
      setStats(mockStats);
    } catch (error) {
      message.error('获取用户统计失败');
    }
  };

  useEffect(() => {
    fetchProfile();
    fetchActivities();
    fetchStats();
  }, []);

  // 处理表单提交
  const handleSubmit = async (values: any) => {
    try {
      setLoading(true);
      await apiClient.updateProfile(values);
      message.success('资料更新成功');
      setEditing(false);
      await fetchProfile();
      updateUser({ ...user, ...values });
    } catch (error) {
      message.error('资料更新失败');
    } finally {
      setLoading(false);
    }
  };

  // 头像上传配置
  const uploadProps: UploadProps = {
    name: 'avatar',
    action: '/api/v1/upload/avatar',
    headers: {
      authorization: `Bearer ${localStorage.getItem('token')}`,
    },
    beforeUpload: (file) => {
      const isJpgOrPng = file.type === 'image/jpeg' || file.type === 'image/png';
      if (!isJpgOrPng) {
        message.error('只能上传 JPG/PNG 格式的图片!');
        return false;
      }
      const isLt2M = file.size / 1024 / 1024 < 2;
      if (!isLt2M) {
        message.error('图片大小不能超过 2MB!');
        return false;
      }
      return true;
    },
    onChange: (info) => {
      if (info.file.status === 'uploading') {
        setUploading(true);
      } else if (info.file.status === 'done') {
        setUploading(false);
        if (info.file.response?.success) {
          const avatarUrl = info.file.response.data.url;
          setProfile(prev => prev ? { ...prev, avatar: avatarUrl } : null);
          updateUser({ ...user, avatar: avatarUrl });
          message.success('头像上传成功');
        } else {
          message.error('头像上传失败');
        }
      } else if (info.file.status === 'error') {
        setUploading(false);
        message.error('头像上传失败');
      }
    },
  };

  // 获取活动类型图标
  const getActivityIcon = (type: string) => {
    switch (type) {
      case 'wisdom':
        return <BookOutlined style={{ color: '#1890ff' }} />;
      case 'comment':
        return <MessageOutlined style={{ color: '#52c41a' }} />;
      case 'like':
        return <TrophyOutlined style={{ color: '#faad14' }} />;
      default:
        return <HistoryOutlined />;
    }
  };

  // 标签页配置
  const tabItems: TabsProps['items'] = [
    {
      key: 'profile',
      label: '基本信息',
      children: (
        <Card>
          <Form
            form={form}
            layout="vertical"
            onFinish={handleSubmit}
            disabled={!editing}
          >
            <Row gutter={24}>
              <Col xs={24} sm={8}>
                <div className="text-center mb-6">
                  <div className="relative inline-block">
                    <Avatar
                      size={120}
                      src={profile?.avatar}
                      icon={<UserOutlined />}
                      className="mb-4"
                    />
                    {editing && (
                      <Upload {...uploadProps} showUploadList={false}>
                        <Button
                          type="primary"
                          shape="circle"
                          icon={<CameraOutlined />}
                          loading={uploading}
                          className="absolute bottom-0 right-0"
                          size="small"
                        />
                      </Upload>
                    )}
                  </div>
                  <div>
                    <Title level={4}>{profile?.name || profile?.username}</Title>
                    <Text type="secondary">{profile?.email}</Text>
                  </div>
                </div>
              </Col>
              
              <Col xs={24} sm={16}>
                <Row gutter={16}>
                  <Col xs={24} sm={12}>
                    <Form.Item
                      name="name"
                      label="姓名"
                      rules={[{ max: 50, message: '姓名不能超过50个字符' }]}
                    >
                      <Input placeholder="请输入姓名" />
                    </Form.Item>
                  </Col>
                  <Col xs={24} sm={12}>
                    <Form.Item
                      name="username"
                      label="用户名"
                      rules={[
                        { required: true, message: '请输入用户名' },
                        { min: 3, max: 20, message: '用户名长度为3-20个字符' }
                      ]}
                    >
                      <Input placeholder="请输入用户名" />
                    </Form.Item>
                  </Col>
                  <Col xs={24} sm={12}>
                    <Form.Item
                      name="email"
                      label="邮箱"
                      rules={[
                        { required: true, message: '请输入邮箱' },
                        { type: 'email', message: '请输入有效的邮箱地址' }
                      ]}
                    >
                      <Input placeholder="请输入邮箱" />
                    </Form.Item>
                  </Col>
                  <Col xs={24} sm={12}>
                    <Form.Item
                      name="phone"
                      label="手机号"
                      rules={[
                        { pattern: /^1[3-9]\d{9}$/, message: '请输入有效的手机号' }
                      ]}
                    >
                      <Input placeholder="请输入手机号" />
                    </Form.Item>
                  </Col>
                  <Col xs={24} sm={12}>
                    <Form.Item
                      name="location"
                      label="所在地"
                    >
                      <Input placeholder="请输入所在地" />
                    </Form.Item>
                  </Col>
                  <Col xs={24} sm={12}>
                    <Form.Item
                      name="website"
                      label="个人网站"
                      rules={[
                        { type: 'url', message: '请输入有效的网址' }
                      ]}
                    >
                      <Input placeholder="请输入个人网站" />
                    </Form.Item>
                  </Col>
                  <Col xs={24}>
                    <Form.Item
                      name="bio"
                      label="个人简介"
                      rules={[{ max: 200, message: '个人简介不能超过200个字符' }]}
                    >
                      <TextArea
                        rows={4}
                        placeholder="介绍一下自己吧..."
                        showCount
                        maxLength={200}
                      />
                    </Form.Item>
                  </Col>
                </Row>
              </Col>
            </Row>

            <Divider />

            <div className="text-center">
              {editing ? (
                <Space>
                  <Button onClick={() => setEditing(false)}>
                    取消
                  </Button>
                  <Button
                    type="primary"
                    htmlType="submit"
                    loading={loading}
                    icon={<SaveOutlined />}
                  >
                    保存
                  </Button>
                </Space>
              ) : (
                <Button
                  type="primary"
                  icon={<EditOutlined />}
                  onClick={() => setEditing(true)}
                >
                  编辑资料
                </Button>
              )}
            </div>
          </Form>
        </Card>
      ),
    },
    {
      key: 'stats',
      label: '数据统计',
      children: (
        <Row gutter={16}>
          <Col xs={24} sm={12} md={6}>
            <Card>
              <Statistic
                title="发布智慧"
                value={stats?.totalWisdom || 0}
                prefix={<BookOutlined />}
                valueStyle={{ color: '#1890ff' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Card>
              <Statistic
                title="评论数"
                value={stats?.totalComments || 0}
                prefix={<MessageOutlined />}
                valueStyle={{ color: '#52c41a' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Card>
              <Statistic
                title="获赞数"
                value={stats?.totalLikes || 0}
                prefix={<TrophyOutlined />}
                valueStyle={{ color: '#faad14' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Card>
              <Statistic
                title="浏览量"
                value={stats?.totalViews || 0}
                prefix={<UserOutlined />}
                valueStyle={{ color: '#722ed1' }}
              />
            </Card>
          </Col>
          <Col xs={24} md={12}>
            <Card title="本月活跃度">
              <div className="mb-4">
                <Text>智慧发布</Text>
                <Progress
                  percent={(stats?.monthlyWisdom || 0) * 10}
                  format={() => `${stats?.monthlyWisdom || 0} 篇`}
                />
              </div>
              <div>
                <Text>评论互动</Text>
                <Progress
                  percent={Math.min((stats?.monthlyComments || 0) * 2, 100)}
                  format={() => `${stats?.monthlyComments || 0} 条`}
                />
              </div>
            </Card>
          </Col>
          <Col xs={24} md={12}>
            <Card title="成就徽章">
              <Space wrap>
                <Tag color="gold" icon={<TrophyOutlined />}>
                  智慧达人
                </Tag>
                <Tag color="blue" icon={<BookOutlined />}>
                  文化学者
                </Tag>
                <Tag color="green" icon={<MessageOutlined />}>
                  活跃评论者
                </Tag>
              </Space>
            </Card>
          </Col>
        </Row>
      ),
    },
    {
      key: 'activity',
      label: '最近活动',
      children: (
        <Card>
          <List
            itemLayout="horizontal"
            dataSource={activities}
            renderItem={(item) => (
              <List.Item>
                <List.Item.Meta
                  avatar={getActivityIcon(item.type)}
                  title={item.title}
                  description={
                    <div>
                      <div>{item.description}</div>
                      <Text type="secondary" className="text-sm">
                        {dayjs(item.createdAt).fromNow()}
                      </Text>
                    </div>
                  }
                />
              </List.Item>
            )}
          />
        </Card>
      ),
    },
  ];

  if (loading && !profile) {
    return (
      <div className="flex justify-center items-center min-h-96">
        <div>加载中...</div>
      </div>
    );
  }

  return (
    <div className="max-w-6xl mx-auto p-6">
      <div className="mb-6">
        <Title level={2}>个人资料</Title>
        <Paragraph type="secondary">
          管理您的个人信息和账户设置
        </Paragraph>
      </div>

      {profile && (
        <Alert
          message="账户信息"
          description={
            <div>
              <Text>注册时间: {dayjs(profile.createdAt).format('YYYY-MM-DD')}</Text>
              <br />
              <Text>最后登录: {profile.lastLogin ? dayjs(profile.lastLogin).format('YYYY-MM-DD HH:mm') : '未知'}</Text>
              <br />
              <Text>登录次数: {profile.loginCount} 次</Text>
            </div>
          }
          type="info"
          showIcon
          className="mb-6"
        />
      )}

      <Tabs items={tabItems} />
    </div>
  );
};

export default Profile;