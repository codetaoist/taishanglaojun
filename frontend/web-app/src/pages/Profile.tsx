import React, { useState, useEffect } from 'react';
import {
  Card,
  Form,
  Input,
  Button,
  Avatar,
  Upload,
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
import { getNotificationInstance } from '../services/notificationService';
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
import { useAuthContext as useAuth } from '../contexts/AuthContext';
import './Profile.css';

// 直接定义User接口，避免导入问题
interface User {
  id: string;
  username: string;
  email: string;
  avatar?: string;
  bio?: string;
  created_at: string;
  updated_at: string;
  roles?: string[];
  permissions?: string[];
}

interface UpdateUserRequest {
  username?: string;
  email?: string;
  avatar?: string;
  bio?: string;
}
import { apiClient } from '../services/api';
import dayjs from 'dayjs';

const { Title, Text, Paragraph } = Typography;
const { TextArea } = Input;

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
  const [profile, setProfile] = useState<User | null>(null);
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
      getNotificationInstance().error({
        message: '错误',
        description: '获取用户资料失败'
      });
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
      getNotificationInstance().error({
        message: '错误',
        description: '获取用户活动失败'
      });
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
      getNotificationInstance().error({
        message: '错误',
        description: '获取用户统计失败'
      });
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
      getNotificationInstance().success({
        message: '成功',
        description: '资料更新成功'
      });
      setEditing(false);
      await fetchProfile();
      updateUser({ ...user, ...values });
    } catch (error) {
      getNotificationInstance().error({
        message: '错误',
        description: '资料更新失败'
      });
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
        getNotificationInstance().error({
          message: '错误',
          description: '只能上传 JPG/PNG 格式的图片!'
        });
        return false;
      }
      const isLt2M = file.size / 1024 / 1024 < 2;
      if (!isLt2M) {
        getNotificationInstance().error({
          message: '错误',
          description: '图片大小不能超过 2MB!'
        });
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
          getNotificationInstance().success({
            message: '成功',
            description: '头像上传成功'
          });
        } else {
          getNotificationInstance().error({
            message: '错误',
            description: '头像上传失败'
          });
        }
      } else if (info.file.status === 'error') {
        setUploading(false);
        getNotificationInstance().error({
          message: '错误',
          description: '头像上传失败'
        });
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
        <Card className="profile-card">
          <Form
            form={form}
            layout="vertical"
            onFinish={handleSubmit}
            disabled={!editing}
            className="profile-form"
          >
            <Row gutter={24}>
              <Col xs={24} sm={8}>
                <div className="profile-avatar-section">
                  <div className="profile-avatar-container">
                    <Avatar
                      size={120}
                      src={profile?.avatar}
                      icon={<UserOutlined />}
                      className="profile-avatar"
                    />
                    {editing && (
                      <Upload {...uploadProps} showUploadList={false}>
                        <Button
                          type="primary"
                          shape="circle"
                          icon={<CameraOutlined />}
                          loading={uploading}
                          className="profile-avatar-upload"
                          size="small"
                        />
                      </Upload>
                    )}
                  </div>
                  <div>
                    <Title level={4} className="profile-user-name">{profile?.name || profile?.username}</Title>
                    <Text className="profile-user-email">{profile?.email}</Text>
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

            <Divider className="profile-divider" />

            <div className="text-center">
              {editing ? (
                <Space>
                  <Button onClick={() => setEditing(false)} className="profile-button">
                    取消
                  </Button>
                  <Button
                    type="primary"
                    htmlType="submit"
                    loading={loading}
                    icon={<SaveOutlined />}
                    className="profile-button-primary"
                  >
                    保存
                  </Button>
                </Space>
              ) : (
                <Button
                  type="primary"
                  icon={<EditOutlined />}
                  onClick={() => setEditing(true)}
                  className="profile-button-primary"
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
            <Card className="profile-stat-card">
              <Statistic
                title="发布智慧"
                value={stats?.totalWisdom || 0}
                prefix={<BookOutlined />}
                valueStyle={{ color: '#1890ff' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Card className="profile-stat-card">
              <Statistic
                title="评论数"
                value={stats?.totalComments || 0}
                prefix={<MessageOutlined />}
                valueStyle={{ color: '#52c41a' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Card className="profile-stat-card">
              <Statistic
                title="获赞数"
                value={stats?.totalLikes || 0}
                prefix={<TrophyOutlined />}
                valueStyle={{ color: '#faad14' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Card className="profile-stat-card">
              <Statistic
                title="浏览量"
                value={stats?.totalViews || 0}
                prefix={<UserOutlined />}
                valueStyle={{ color: '#722ed1' }}
              />
            </Card>
          </Col>
          <Col xs={24} md={12}>
            <Card title="本月活跃度" className="profile-card">
              <div className="mb-4">
                <Text className="profile-text">智慧发布</Text>
                <Progress
                  percent={(stats?.monthlyWisdom || 0) * 10}
                  format={() => `${stats?.monthlyWisdom || 0} 篇`}
                  className="profile-progress"
                />
              </div>
              <div>
                <Text className="profile-text">评论互动</Text>
                <Progress
                  percent={Math.min((stats?.monthlyComments || 0) * 2, 100)}
                  format={() => `${stats?.monthlyComments || 0} 条`}
                  className="profile-progress"
                />
              </div>
            </Card>
          </Col>
          <Col xs={24} md={12}>
            <Card title="成就徽章" className="profile-card">
              <Space wrap>
                <Tag icon={<TrophyOutlined />} className="profile-tag">
                  智慧达人
                </Tag>
                <Tag icon={<BookOutlined />} className="profile-tag">
                  文化学者
                </Tag>
                <Tag icon={<MessageOutlined />} className="profile-tag">
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
        <Card className="profile-card">
          <List
            itemLayout="horizontal"
            dataSource={activities}
            className="profile-list"
            renderItem={(item) => (
              <List.Item>
                <List.Item.Meta
                  avatar={<div className="profile-activity-icon">{getActivityIcon(item.type)}</div>}
                  title={item.title}
                  description={
                    <div>
                      <div className="profile-text">{item.description}</div>
                      <Text className="profile-text-secondary">
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
    <div className="profile-container max-w-6xl mx-auto p-6">
      <div className="profile-header mb-6">
        <Title level={2} className="profile-title">个人资料</Title>
        <Paragraph type="secondary" className="profile-description">
          管理您的个人信息和账户设置
        </Paragraph>
      </div>

      {profile && (
        <Alert
          message="账户信息"
          description="您的账户信息由系统自动管理，部分信息可能无法修改。"
          type="info"
          showIcon
          className="profile-alert"
        />
      )}

      <Tabs
        defaultActiveKey="profile"
        items={tabItems}
        className="profile-tabs"
      />
    </div>
  );
};

export default Profile;