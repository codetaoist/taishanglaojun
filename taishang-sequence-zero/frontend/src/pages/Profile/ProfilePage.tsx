import React, { useState, useEffect } from 'react';
import { Row, Col, Card, Avatar, Button, Form, Input, Select, Switch, Divider, Upload, message, Tabs, List, Tag, Progress, Modal, Alert } from 'antd';
import {
  UserOutlined,
  EditOutlined,
  SettingOutlined,
  CameraOutlined,
  SaveOutlined,
  LockOutlined,
  BellOutlined,
  GlobalOutlined,
  EyeOutlined,
  SafetyOutlined,
  TrophyOutlined,
  StarOutlined,
  HeartOutlined,
  BookOutlined,
  HistoryOutlined,
  DeleteOutlined,
} from '@ant-design/icons';
import { useAppSelector, useAppDispatch } from '../../hooks/redux';
import { updateUserProfile, changePassword } from '../../store/slices/authSlice';
import { setTheme, setLanguage } from '../../store/slices/uiSlice';

const { Option } = Select;
const { TabPane } = Tabs;
const { TextArea } = Input;

interface ProfileFormData {
  username: string;
  email: string;
  fullName: string;
  bio: string;
  location: string;
  website: string;
}

interface PasswordFormData {
  currentPassword: string;
  newPassword: string;
  confirmPassword: string;
}

interface PreferencesData {
  theme: string;
  language: string;
  notifications: {
    email: boolean;
    push: boolean;
    wisdom: boolean;
    consciousness: boolean;
  };
  privacy: {
    profileVisible: boolean;
    activityVisible: boolean;
    progressVisible: boolean;
  };
}

const ProfilePage: React.FC = () => {
  const [profileForm] = Form.useForm();
  const [passwordForm] = Form.useForm();
  const [preferencesForm] = Form.useForm();
  
  const dispatch = useAppDispatch();
  const { user, loading } = useAppSelector(state => state.auth);
  const { theme, language } = useAppSelector(state => state.ui);
  const { learningProgress } = useAppSelector(state => state.cultural);
  const { currentEntity, sessions } = useAppSelector(state => state.consciousness || { currentEntity: null, sessions: [] });

  const [activeTab, setActiveTab] = useState('profile');
  const [editMode, setEditMode] = useState(false);
  const [avatarUrl, setAvatarUrl] = useState(user?.avatar || '');
  const [deleteAccountModal, setDeleteAccountModal] = useState(false);

  // 初始化表单数据
  useEffect(() => {
    if (user) {
      profileForm.setFieldsValue({
        username: user.username,
        email: user.email,
        fullName: user.fullName || '',
        bio: user.bio || '',
        location: user.location || '',
        website: user.website || '',
      });
    }
  }, [user, profileForm]);

  // 获取文本内容
  const getText = (zhText: string, enText: string) => {
    return language === 'en-US' ? enText : zhText;
  };

  // 处理头像上传
  const handleAvatarUpload = (info: any) => {
    if (info.file.status === 'uploading') {
      return;
    }
    if (info.file.status === 'done') {
      // 获取上传后的URL
      const url = info.file.response?.url || URL.createObjectURL(info.file.originFileObj);
      setAvatarUrl(url);
      message.success(getText('头像上传成功', 'Avatar uploaded successfully'));
    } else if (info.file.status === 'error') {
      message.error(getText('头像上传失败', 'Avatar upload failed'));
    }
  };

  // 保存个人资料
  const handleSaveProfile = async (values: ProfileFormData) => {
    try {
      await dispatch(updateUserProfile({
        ...values,
        avatar: avatarUrl,
      })).unwrap();
      setEditMode(false);
      message.success(getText('个人资料更新成功', 'Profile updated successfully'));
    } catch (error) {
      message.error(getText('个人资料更新失败', 'Failed to update profile'));
    }
  };

  // 修改密码
  const handleChangePassword = async (values: PasswordFormData) => {
    try {
      await dispatch(changePassword({
        currentPassword: values.currentPassword,
        newPassword: values.newPassword,
      })).unwrap();
      passwordForm.resetFields();
      message.success(getText('密码修改成功', 'Password changed successfully'));
    } catch (error) {
      message.error(getText('密码修改失败', 'Failed to change password'));
    }
  };

  // 保存偏好设置
  const handleSavePreferences = async (values: PreferencesData) => {
    try {
      // 更新UI偏好设置
       dispatch(setTheme(values.theme as 'light' | 'dark' | 'auto'));
       dispatch(setLanguage(values.language as 'zh-CN' | 'en-US'));
      message.success(getText('偏好设置保存成功', 'Preferences saved successfully'));
    } catch (error) {
      message.error(getText('偏好设置保存失败', 'Failed to save preferences'));
    }
  };

  // 删除账户
  const handleDeleteAccount = () => {
    // 这里应该调用删除账户的API
    message.success(getText('账户删除请求已提交', 'Account deletion request submitted'));
    setDeleteAccountModal(false);
  };

  // 模拟成就数据
  const achievements = [
    {
      id: 1,
      title: getText('智慧探索者', 'Wisdom Explorer'),
      description: getText('完成10个智慧学习任务', 'Complete 10 wisdom learning tasks'),
      icon: <BookOutlined />,
      earned: true,
      earnedDate: '2024-01-15',
    },
    {
      id: 2,
      title: getText('意识融合大师', 'Consciousness Fusion Master'),
      description: getText('进行50次意识融合会话', 'Complete 50 consciousness fusion sessions'),
      icon: <HeartOutlined />,
      earned: true,
      earnedDate: '2024-01-20',
    },
    {
      id: 3,
      title: getText('文化传承者', 'Cultural Inheritor'),
      description: getText('学习完成5个文化主题', 'Complete 5 cultural themes'),
      icon: <TrophyOutlined />,
      earned: false,
      progress: 60,
    },
  ];

  // 模拟活动历史
  const recentActivities = [
    {
      id: 1,
      type: 'consciousness',
      title: getText('完成意识融合会话', 'Completed consciousness fusion session'),
      time: '2小时前',
      icon: <HeartOutlined style={{ color: '#ff4d4f' }} />,
    },
    {
      id: 2,
      type: 'cultural',
      title: getText('学习《道德经》第三章', 'Studied Tao Te Ching Chapter 3'),
      time: '5小时前',
      icon: <BookOutlined style={{ color: '#1890ff' }} />,
    },
    {
      id: 3,
      type: 'achievement',
      title: getText('获得"智慧探索者"成就', 'Earned "Wisdom Explorer" achievement'),
      time: '1天前',
      icon: <TrophyOutlined style={{ color: '#faad14' }} />,
    },
  ];

  return (
    <div className="profile-page">
      {/* 页面标题 */}
      <div className="page-header">
        <h1 className="page-title">
          <UserOutlined style={{ marginRight: 12, color: '#1890ff' }} />
          {getText('个人资料', 'Profile')}
        </h1>
        <p className="page-description">
          {getText(
            '管理您的个人信息、偏好设置和隐私选项',
            'Manage your personal information, preferences, and privacy settings'
          )}
        </p>
      </div>

      <Row gutter={[24, 24]}>
        {/* 左侧个人信息卡片 */}
        <Col xs={24} lg={8}>
          <Card className="profile-card">
            <div className="profile-header">
              <div className="avatar-section">
                <Avatar
                  size={100}
                  src={avatarUrl}
                  icon={<UserOutlined />}
                  className="profile-avatar"
                />
                {editMode && (
                  <Upload
                    name="avatar"
                    listType="picture"
                    className="avatar-uploader"
                    showUploadList={false}
                    action="/api/upload/avatar"
                    beforeUpload={(file) => {
                      const isJpgOrPng = file.type === 'image/jpeg' || file.type === 'image/png';
                      if (!isJpgOrPng) {
                        message.error(getText('只能上传JPG/PNG格式的图片', 'You can only upload JPG/PNG files'));
                      }
                      const isLt2M = file.size / 1024 / 1024 < 2;
                      if (!isLt2M) {
                        message.error(getText('图片大小不能超过2MB', 'Image must be smaller than 2MB'));
                      }
                      return isJpgOrPng && isLt2M;
                    }}
                    onChange={handleAvatarUpload}
                  >
                    <Button
                      type="primary"
                      shape="circle"
                      icon={<CameraOutlined />}
                      className="avatar-upload-btn"
                    />
                  </Upload>
                )}
              </div>
              
              <div className="profile-info">
                <h2 className="profile-name">{user?.fullName || user?.username}</h2>
                <p className="profile-email">{user?.email}</p>
                {user?.bio && <p className="profile-bio">{user.bio}</p>}
              </div>
              
              <Button
                type={editMode ? 'default' : 'primary'}
                icon={editMode ? <SaveOutlined /> : <EditOutlined />}
                onClick={() => setEditMode(!editMode)}
                className="edit-button"
              >
                {editMode ? getText('取消编辑', 'Cancel') : getText('编辑资料', 'Edit Profile')}
              </Button>
            </div>
            
            {/* 统计信息 */}
            <Divider />
            <div className="profile-stats">
              <div className="stat-item">
                <div className="stat-number">{sessions?.length || 0}</div>
                <div className="stat-label">{getText('融合会话', 'Fusion Sessions')}</div>
              </div>
              <div className="stat-item">
                <div className="stat-number">{(learningProgress as any)?.studiedItems || 0}</div>
                <div className="stat-label">{getText('学习项目', 'Studied Items')}</div>
              </div>
              <div className="stat-item">
                <div className="stat-number">{achievements.filter(a => a.earned).length}</div>
                <div className="stat-label">{getText('获得成就', 'Achievements')}</div>
              </div>
            </div>
          </Card>
        </Col>

        {/* 右侧详细信息 */}
        <Col xs={24} lg={16}>
          <Card className="details-card">
            <Tabs activeKey={activeTab} onChange={setActiveTab} size="large">
              {/* 个人资料 */}
              <TabPane 
                tab={
                  <span>
                    <UserOutlined />
                    {getText('个人资料', 'Profile')}
                  </span>
                } 
                key="profile"
              >
                <Form
                  form={profileForm}
                  layout="vertical"
                  onFinish={handleSaveProfile}
                  disabled={!editMode}
                >
                  <Row gutter={16}>
                    <Col xs={24} sm={12}>
                      <Form.Item
                        name="username"
                        label={getText('用户名', 'Username')}
                        rules={[
                          { required: true, message: getText('请输入用户名', 'Please enter username') },
                          { min: 3, message: getText('用户名至少3位字符', 'Username must be at least 3 characters') },
                        ]}
                      >
                        <Input disabled />
                      </Form.Item>
                    </Col>
                    <Col xs={24} sm={12}>
                      <Form.Item
                        name="email"
                        label={getText('邮箱地址', 'Email Address')}
                        rules={[
                          { required: true, message: getText('请输入邮箱地址', 'Please enter email address') },
                          { type: 'email', message: getText('请输入有效的邮箱地址', 'Please enter a valid email address') },
                        ]}
                      >
                        <Input disabled />
                      </Form.Item>
                    </Col>
                  </Row>
                  
                  <Row gutter={16}>
                    <Col xs={24} sm={12}>
                      <Form.Item
                        name="fullName"
                        label={getText('真实姓名', 'Full Name')}
                      >
                        <Input placeholder={getText('请输入真实姓名', 'Enter your full name')} />
                      </Form.Item>
                    </Col>
                    <Col xs={24} sm={12}>
                      <Form.Item
                        name="location"
                        label={getText('所在地区', 'Location')}
                      >
                        <Input placeholder={getText('请输入所在地区', 'Enter your location')} />
                      </Form.Item>
                    </Col>
                  </Row>
                  
                  <Form.Item
                    name="website"
                    label={getText('个人网站', 'Website')}
                  >
                    <Input placeholder={getText('请输入个人网站', 'Enter your website')} />
                  </Form.Item>
                  
                  <Form.Item
                    name="bio"
                    label={getText('个人简介', 'Bio')}
                  >
                    <TextArea 
                      rows={4} 
                      placeholder={getText('介绍一下自己...', 'Tell us about yourself...')} 
                      maxLength={200}
                      showCount
                    />
                  </Form.Item>
                  
                  {editMode && (
                    <Form.Item>
                      <Button type="primary" htmlType="submit" loading={loading} size="large">
                        {getText('保存更改', 'Save Changes')}
                      </Button>
                    </Form.Item>
                  )}
                </Form>
              </TabPane>

              {/* 安全设置 */}
              <TabPane 
                tab={
                  <span>
                    <LockOutlined />
                    {getText('安全设置', 'Security')}
                  </span>
                } 
                key="security"
              >
                <div className="security-section">
                  <h3>{getText('修改密码', 'Change Password')}</h3>
                  <Form
                    form={passwordForm}
                    layout="vertical"
                    onFinish={handleChangePassword}
                  >
                    <Form.Item
                      name="currentPassword"
                      label={getText('当前密码', 'Current Password')}
                      rules={[
                        { required: true, message: getText('请输入当前密码', 'Please enter current password') },
                      ]}
                    >
                      <Input.Password placeholder={getText('请输入当前密码', 'Enter current password')} />
                    </Form.Item>
                    
                    <Form.Item
                      name="newPassword"
                      label={getText('新密码', 'New Password')}
                      rules={[
                        { required: true, message: getText('请输入新密码', 'Please enter new password') },
                        { min: 6, message: getText('密码至少6位字符', 'Password must be at least 6 characters') },
                      ]}
                    >
                      <Input.Password placeholder={getText('请输入新密码', 'Enter new password')} />
                    </Form.Item>
                    
                    <Form.Item
                      name="confirmPassword"
                      label={getText('确认新密码', 'Confirm New Password')}
                      dependencies={['newPassword']}
                      rules={[
                        { required: true, message: getText('请确认新密码', 'Please confirm new password') },
                        ({ getFieldValue }) => ({
                          validator(_, value) {
                            if (!value || getFieldValue('newPassword') === value) {
                              return Promise.resolve();
                            }
                            return Promise.reject(new Error(getText('两次输入的密码不一致', 'Passwords do not match')));
                          },
                        }),
                      ]}
                    >
                      <Input.Password placeholder={getText('请确认新密码', 'Confirm new password')} />
                    </Form.Item>
                    
                    <Form.Item>
                      <Button type="primary" htmlType="submit" loading={loading}>
                        {getText('修改密码', 'Change Password')}
                      </Button>
                    </Form.Item>
                  </Form>
                  
                  <Divider />
                  
                  <div className="danger-zone">
                    <h3 className="danger-title">{getText('危险操作', 'Danger Zone')}</h3>
                    <Alert
                      message={getText('删除账户', 'Delete Account')}
                      description={getText(
                        '删除账户将永久删除您的所有数据，此操作不可恢复',
                        'Deleting your account will permanently remove all your data. This action cannot be undone.'
                      )}
                      type="warning"
                      showIcon
                      action={
                        <Button 
                          danger 
                          size="small" 
                          onClick={() => setDeleteAccountModal(true)}
                        >
                          {getText('删除账户', 'Delete Account')}
                        </Button>
                      }
                    />
                  </div>
                </div>
              </TabPane>

              {/* 偏好设置 */}
              <TabPane 
                tab={
                  <span>
                    <SettingOutlined />
                    {getText('偏好设置', 'Preferences')}
                  </span>
                } 
                key="preferences"
              >
                <Form
                  form={preferencesForm}
                  layout="vertical"
                  onFinish={handleSavePreferences}
                  initialValues={{
                    theme,
                    language,
                    notifications: {
                      email: true,
                      push: true,
                      wisdom: true,
                      consciousness: true,
                    },
                    privacy: {
                      profileVisible: true,
                      activityVisible: true,
                      progressVisible: true,
                    },
                  }}
                >
                  <div className="preference-section">
                    <h3>{getText('界面设置', 'Interface Settings')}</h3>
                    <Row gutter={16}>
                      <Col xs={24} sm={12}>
                        <Form.Item
                          name="theme"
                          label={getText('主题模式', 'Theme Mode')}
                        >
                          <Select>
                            <Option value="light">{getText('浅色模式', 'Light Mode')}</Option>
                            <Option value="dark">{getText('深色模式', 'Dark Mode')}</Option>
                            <Option value="auto">{getText('跟随系统', 'Follow System')}</Option>
                          </Select>
                        </Form.Item>
                      </Col>
                      <Col xs={24} sm={12}>
                        <Form.Item
                          name="language"
                          label={getText('语言设置', 'Language')}
                        >
                          <Select>
                            <Option value="zh">中文</Option>
                            <Option value="en">English</Option>
                          </Select>
                        </Form.Item>
                      </Col>
                    </Row>
                  </div>
                  
                  <Divider />
                  
                  <div className="preference-section">
                    <h3>{getText('通知设置', 'Notification Settings')}</h3>
                    <Form.Item name={['notifications', 'email']} valuePropName="checked">
                      <div className="preference-item">
                        <div className="preference-info">
                          <BellOutlined className="preference-icon" />
                          <div>
                            <div className="preference-title">{getText('邮件通知', 'Email Notifications')}</div>
                            <div className="preference-desc">{getText('接收重要更新和提醒', 'Receive important updates and reminders')}</div>
                          </div>
                        </div>
                        <Switch />
                      </div>
                    </Form.Item>
                    
                    <Form.Item name={['notifications', 'wisdom']} valuePropName="checked">
                      <div className="preference-item">
                        <div className="preference-info">
                          <BookOutlined className="preference-icon" />
                          <div>
                            <div className="preference-title">{getText('智慧提醒', 'Wisdom Reminders')}</div>
                            <div className="preference-desc">{getText('每日智慧内容推送', 'Daily wisdom content push')}</div>
                          </div>
                        </div>
                        <Switch />
                      </div>
                    </Form.Item>
                    
                    <Form.Item name={['notifications', 'consciousness']} valuePropName="checked">
                      <div className="preference-item">
                        <div className="preference-info">
                          <HeartOutlined className="preference-icon" />
                          <div>
                            <div className="preference-title">{getText('融合提醒', 'Fusion Reminders')}</div>
                            <div className="preference-desc">{getText('意识融合会话提醒', 'Consciousness fusion session reminders')}</div>
                          </div>
                        </div>
                        <Switch />
                      </div>
                    </Form.Item>
                  </div>
                  
                  <Divider />
                  
                  <div className="preference-section">
                    <h3>{getText('隐私设置', 'Privacy Settings')}</h3>
                    <Form.Item name={['privacy', 'profileVisible']} valuePropName="checked">
                      <div className="preference-item">
                        <div className="preference-info">
                          <EyeOutlined className="preference-icon" />
                          <div>
                            <div className="preference-title">{getText('公开个人资料', 'Public Profile')}</div>
                            <div className="preference-desc">{getText('允许其他用户查看您的个人资料', 'Allow other users to view your profile')}</div>
                          </div>
                        </div>
                        <Switch />
                      </div>
                    </Form.Item>
                    
                    <Form.Item name={['privacy', 'activityVisible']} valuePropName="checked">
                      <div className="preference-item">
                        <div className="preference-info">
                          <HistoryOutlined className="preference-icon" />
                          <div>
                            <div className="preference-title">{getText('公开活动记录', 'Public Activity')}</div>
                            <div className="preference-desc">{getText('显示您的学习活动记录', 'Show your learning activity history')}</div>
                          </div>
                        </div>
                        <Switch />
                      </div>
                    </Form.Item>
                  </div>
                  
                  <Form.Item>
                    <Button type="primary" htmlType="submit" loading={loading} size="large">
                      {getText('保存设置', 'Save Settings')}
                    </Button>
                  </Form.Item>
                </Form>
              </TabPane>

              {/* 成就与活动 */}
              <TabPane 
                tab={
                  <span>
                    <TrophyOutlined />
                    {getText('成就活动', 'Achievements')}
                  </span>
                } 
                key="achievements"
              >
                <div className="achievements-section">
                  <h3>{getText('我的成就', 'My Achievements')}</h3>
                  <List
                    grid={{ gutter: 16, xs: 1, sm: 2, md: 2, lg: 3 }}
                    dataSource={achievements}
                    renderItem={achievement => (
                      <List.Item>
                        <Card 
                          className={`achievement-card ${achievement.earned ? 'earned' : 'not-earned'}`}
                          hoverable
                        >
                          <div className="achievement-content">
                            <div className="achievement-icon">
                              {achievement.icon}
                            </div>
                            <div className="achievement-info">
                              <h4 className="achievement-title">{achievement.title}</h4>
                              <p className="achievement-desc">{achievement.description}</p>
                              {achievement.earned ? (
                                <Tag color="green">
                                  {getText('已获得', 'Earned')} - {achievement.earnedDate}
                                </Tag>
                              ) : (
                                <div className="achievement-progress">
                                  <Progress percent={achievement.progress} size="small" />
                                  <span className="progress-text">{achievement.progress}%</span>
                                </div>
                              )}
                            </div>
                          </div>
                        </Card>
                      </List.Item>
                    )}
                  />
                  
                  <Divider />
                  
                  <h3>{getText('最近活动', 'Recent Activities')}</h3>
                  <List
                    dataSource={recentActivities}
                    renderItem={activity => (
                      <List.Item>
                        <List.Item.Meta
                          avatar={<Avatar icon={activity.icon} />}
                          title={activity.title}
                          description={activity.time}
                        />
                      </List.Item>
                    )}
                  />
                </div>
              </TabPane>
            </Tabs>
          </Card>
        </Col>
      </Row>

      {/* 删除账户确认模态框 */}
      <Modal
        title={getText('确认删除账户', 'Confirm Account Deletion')}
        open={deleteAccountModal}
        onCancel={() => setDeleteAccountModal(false)}
        footer={[
          <Button key="cancel" onClick={() => setDeleteAccountModal(false)}>
            {getText('取消', 'Cancel')}
          </Button>,
          <Button key="delete" type="primary" danger onClick={handleDeleteAccount}>
            {getText('确认删除', 'Confirm Delete')}
          </Button>,
        ]}
      >
        <Alert
          message={getText('警告', 'Warning')}
          description={getText(
            '删除账户将永久删除您的所有数据，包括学习进度、融合记录、成就等。此操作不可恢复，请谨慎操作。',
            'Deleting your account will permanently remove all your data, including learning progress, fusion records, achievements, etc. This action cannot be undone, please proceed with caution.'
          )}
          type="warning"
          showIcon
          style={{ marginBottom: 16 }}
        />
        <p>{getText('请输入您的用户名以确认删除：', 'Please enter your username to confirm deletion:')}</p>
        <Input placeholder={user?.username} />
      </Modal>

      <style>{`
        .profile-page {
          padding: 0;
        }

        .page-header {
          margin-bottom: 24px;
        }

        .page-title {
          font-size: 28px;
          font-weight: 600;
          color: var(--text-primary);
          margin-bottom: 8px;
          display: flex;
          align-items: center;
        }

        .page-description {
          font-size: 16px;
          color: var(--text-secondary);
          margin: 0;
          line-height: 1.6;
        }

        .profile-card {
          height: fit-content;
        }

        .profile-header {
          text-align: center;
        }

        .avatar-section {
          position: relative;
          display: inline-block;
          margin-bottom: 16px;
        }

        .profile-avatar {
          border: 3px solid var(--border-light);
        }

        .avatar-uploader {
          position: absolute;
          bottom: 0;
          right: 0;
        }

        .avatar-upload-btn {
          width: 32px;
          height: 32px;
          min-width: 32px;
        }

        .profile-info {
          margin-bottom: 20px;
        }

        .profile-name {
          font-size: 20px;
          font-weight: 600;
          color: var(--text-primary);
          margin-bottom: 4px;
        }

        .profile-email {
          color: var(--text-secondary);
          margin-bottom: 8px;
        }

        .profile-bio {
          color: var(--text-tertiary);
          font-size: 14px;
          line-height: 1.5;
          margin: 0;
        }

        .edit-button {
          width: 100%;
        }

        .profile-stats {
          display: flex;
          justify-content: space-around;
          text-align: center;
        }

        .stat-item {
          flex: 1;
        }

        .stat-number {
          font-size: 20px;
          font-weight: 600;
          color: var(--text-primary);
          line-height: 1;
        }

        .stat-label {
          font-size: 12px;
          color: var(--text-secondary);
          margin-top: 4px;
        }

        .details-card {
          min-height: 600px;
        }

        .security-section {
          max-width: 500px;
        }

        .danger-zone {
          margin-top: 40px;
        }

        .danger-title {
          color: #ff4d4f;
          margin-bottom: 16px;
        }

        .preference-section {
          margin-bottom: 32px;
        }

        .preference-section h3 {
          margin-bottom: 20px;
          color: var(--text-primary);
        }

        .preference-item {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 16px 0;
          border-bottom: 1px solid var(--border-light);
        }

        .preference-item:last-child {
          border-bottom: none;
        }

        .preference-info {
          display: flex;
          align-items: center;
          gap: 12px;
          flex: 1;
        }

        .preference-icon {
          font-size: 16px;
          color: var(--text-secondary);
        }

        .preference-title {
          font-weight: 500;
          color: var(--text-primary);
          margin-bottom: 4px;
        }

        .preference-desc {
          font-size: 12px;
          color: var(--text-tertiary);
          margin: 0;
        }

        .achievements-section h3 {
          margin-bottom: 20px;
          color: var(--text-primary);
        }

        .achievement-card {
          height: 100%;
          transition: all 0.3s;
        }

        .achievement-card.earned {
          border-color: #52c41a;
          background: linear-gradient(135deg, #f6ffed 0%, #f0f9ff 100%);
        }

        .achievement-card.not-earned {
          opacity: 0.6;
        }

        .achievement-content {
          display: flex;
          align-items: flex-start;
          gap: 12px;
        }

        .achievement-icon {
          font-size: 24px;
          color: var(--primary-color);
          margin-top: 4px;
        }

        .achievement-info {
          flex: 1;
        }

        .achievement-title {
          font-size: 16px;
          font-weight: 600;
          color: var(--text-primary);
          margin-bottom: 8px;
        }

        .achievement-desc {
          font-size: 14px;
          color: var(--text-secondary);
          margin-bottom: 12px;
          line-height: 1.4;
        }

        .achievement-progress {
          display: flex;
          align-items: center;
          gap: 8px;
        }

        .progress-text {
          font-size: 12px;
          color: var(--text-tertiary);
          min-width: 30px;
        }

        /* 响应式设计 */
        @media (max-width: 768px) {
          .page-title {
            font-size: 24px;
          }

          .page-description {
            font-size: 14px;
          }

          .profile-stats {
            flex-direction: column;
            gap: 16px;
          }

          .stat-item {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 0 16px;
          }

          .preference-item {
            flex-direction: column;
            align-items: flex-start;
            gap: 12px;
          }

          .achievement-content {
            flex-direction: column;
            text-align: center;
          }
        }
      `}</style>
    </div>
  );
};

export default ProfilePage;