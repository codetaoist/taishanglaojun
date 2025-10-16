import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { 
  Layout, 
  Card, 
  List, 
  Avatar, 
  Button, 
  Tag, 
  Space, 
  Typography, 
  Row, 
  Col,
  Tabs,
  Input,
  Select,
  Divider,
  Badge,
  Tooltip,
  message,
  Modal,
  Form
} from 'antd';
import { 
  UserOutlined, 
  MessageOutlined, 
  HeartOutlined, 
  ShareAltOutlined,
  EyeOutlined,
  PlusOutlined,
  FireOutlined,
  ClockCircleOutlined,
  StarOutlined,
  CommentOutlined,
  LikeOutlined,
  BookOutlined,
  TeamOutlined
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';

const { Content } = Layout;
const { Title, Paragraph, Text } = Typography;
const { TextArea } = Input;
const { Option } = Select;

interface CommunityPost {
  id: string;
  title: string;
  content: string;
  author: {
    id: string;
    username: string;
    avatar?: string;
    level: number;
    title: string;
  };
  category: string;
  tags: string[];
  createdAt: string;
  updatedAt: string;
  likes: number;
  comments: number;
  views: number;
  isLiked: boolean;
  isBookmarked: boolean;
}

interface CommunityUser {
  id: string;
  username: string;
  avatar?: string;
  level: number;
  title: string;
  posts: number;
  followers: number;
  following: number;
  joinDate: string;
}

const Community: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState('posts');
  const [posts, setPosts] = useState<CommunityPost[]>([]);
  const [users, setUsers] = useState<CommunityUser[]>([]);
  const [loading, setLoading] = useState(false);
  const [showNewPost, setShowNewPost] = useState(false);
  const [newPostForm] = Form.useForm();

  // 模拟数据
  useEffect(() => {
    const mockPosts: CommunityPost[] = [
      {
        id: '1',
        title: '《道德经》第一章的现代理解',
        content: '道可道，非常道。名可名，非常名。这句话在现代社会中如何理解？认为...',
        author: {
          id: '1',
          username: '修道者',
          level: 5,
          title: '道学研究者'
        },
        category: '道学讨论',
        tags: ['道德经', '哲学', '修行'],
        createdAt: '2024-01-15T10:30:00Z',
        updatedAt: '2024-01-15T10:30:00Z',
        likes: 42,
        comments: 18,
        views: 256,
        isLiked: false,
        isBookmarked: true
      },
      {
        id: '2',
        title: '分享我的静坐修行心得',
        content: '经过三个月的静坐练习，我有了一些体会想和大家分享...',
        author: {
          id: '2',
          username: '静心居士',
          level: 3,
          title: '修行新手'
        },
        category: '修行心得',
        tags: ['静坐', '冥想', '心得'],
        createdAt: '2024-01-14T15:20:00Z',
        updatedAt: '2024-01-14T15:20:00Z',
        likes: 28,
        comments: 12,
        views: 189,
        isLiked: true,
        isBookmarked: false
      },
      {
        id: '3',
        title: '儒家思想在现代管理中的应用',
        content: '仁义礼智信这五常在现代企业管理中仍然有重要意义...',
        author: {
          id: '3',
          username: '商道智者',
          level: 7,
          title: '儒学导师'
        },
        category: '儒学应用',
        tags: ['儒家', '管理', '商道'],
        createdAt: '2024-01-13T09:15:00Z',
        updatedAt: '2024-01-13T09:15:00Z',
        likes: 67,
        comments: 25,
        views: 432,
        isLiked: false,
        isBookmarked: false
      }
    ];

    const mockUsers: CommunityUser[] = [
      {
        id: '1',
        username: '修道者',
        level: 5,
        title: '道学研究者',
        posts: 23,
        followers: 156,
        following: 89,
        joinDate: '2023-06-15'
      },
      {
        id: '2',
        username: '静心居士',
        level: 3,
        title: '修行新手',
        posts: 8,
        followers: 45,
        following: 67,
        joinDate: '2023-10-20'
      },
      {
        id: '3',
        username: '商道智者',
        level: 7,
        title: '儒学导师',
        posts: 45,
        followers: 289,
        following: 123,
        joinDate: '2023-03-10'
      }
    ];

    setPosts(mockPosts);
    setUsers(mockUsers);
  }, []);

  const handleLike = (postId: string) => {
    setPosts(prev => prev.map(post => 
      post.id === postId 
        ? { 
            ...post, 
            isLiked: !post.isLiked,
            likes: post.isLiked ? post.likes - 1 : post.likes + 1
          }
        : post
    ));
  };

  const handleBookmark = (postId: string) => {
    setPosts(prev => prev.map(post => 
      post.id === postId 
        ? { ...post, isBookmarked: !post.isBookmarked }
        : post
    ));
    message.success(t('community.messages.bookmarkAdded'));
  };

  const getLevelColor = (level: number) => {
    if (level >= 7) return '#f50';
    if (level >= 5) return '#fa8c16';
    if (level >= 3) return '#faad14';
    return '#52c41a';
  };

  const formatTime = (dateString: string) => {
    const date = new Date(dateString);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const days = Math.floor(diff / (1000 * 60 * 60 * 24));
    
    if (days === 0) return t('community.time.today');
    if (days === 1) return t('community.time.yesterday');
    if (days < 7) return t('community.time.daysAgo', { count: days });
    return date.toLocaleDateString();
  };

  const handleSubmitPost = async (values: any) => {
    try {
      setLoading(true);
      
      // 创建新帖子对象
      const newPost: CommunityPost = {
        id: Date.now().toString(),
        title: values.title,
        content: values.content,
        author: {
          id: 'current-user',
          username: '当前用户',
          level: 1,
          title: '新手修行者'
        },
        category: values.category,
        tags: values.tags || [],
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        likes: 0,
        comments: 0,
        views: 0,
        isLiked: false,
        isBookmarked: false
      };

      // 添加到帖子列表开头
      setPosts(prev => [newPost, ...prev]);
      
      // 重置表单并关闭模态框
      newPostForm.resetFields();
      setShowNewPost(false);
      message.success(t('community.messages.postPublishSuccess'));
      
    } catch (error) {
      message.error(t('community.messages.postPublishFailed'));
    } finally {
      setLoading(false);
    }
  };

  const handleComment = (postId: string) => {
    // TODO: 实现评论功能
    message.info(t('community.messages.commentFeatureWip'));
  };

  const handleShare = (postId: string) => {
    // TODO: 实现分享功能
    navigator.clipboard.writeText(`${window.location.origin}/community/post/${postId}`);
    message.success(t('community.messages.linkCopied'));
  };

  const tabItems = [
    {
      key: 'posts',
      label: (
        <span>
          <MessageOutlined />
          {t('community.tabs.posts')}
        </span>
      ),
      children: (
        <div>
          <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
            <Col span={12}>
              <Input.Search 
                placeholder={t('community.search.placeholder')} 
                onSearch={(value) => console.log('搜索:', value)}
              />
            </Col>
            <Col span={6}>
              <Select placeholder="选择分类" style={{ width: '100%' }}>
                <Option value="all">{t('community.categories.all')}</Option>
                <Option value="道学讨论">{t('community.categories.taoismDiscussion')}</Option>
                <Option value="修行心得">{t('community.categories.practiceInsights')}</Option>
                <Option value="儒学应用">{t('community.categories.confucianApplications')}</Option>
                <Option value="佛学智慧">{t('community.categories.buddhistWisdom')}</Option>
              </Select>
            </Col>
            <Col span={6}>
              <Button 
                type="primary" 
                icon={<PlusOutlined />}
                onClick={() => setShowNewPost(true)}
                style={{ width: '100%' }}
              >
                {t('community.actions.postNew')}
              </Button>
            </Col>
          </Row>

          <List
            itemLayout="vertical"
            size="large"
            dataSource={posts}
            renderItem={(post) => (
              <List.Item
                key={post.id}
                actions={[
                  <Space key="like">
                    <Button 
                      type="text" 
                      icon={<LikeOutlined />}
                      onClick={() => handleLike(post.id)}
                      style={{ color: post.isLiked ? '#1890ff' : undefined }}
                    >
                      {post.likes}
                    </Button>
                  </Space>,
                  <Space key="comment">
                    <Button 
                      type="text" 
                      icon={<CommentOutlined />}
                      onClick={() => handleComment(post.id)}
                    >
                      {post.comments}
                    </Button>
                  </Space>,
                  <Space key="view">
                    <Button type="text" icon={<EyeOutlined />}>
                      {post.views}
                    </Button>
                  </Space>,
                  <Space key="share">
                    <Button 
                      type="text" 
                      icon={<ShareAltOutlined />}
                      onClick={() => handleShare(post.id)}
                    >
                      分享
                    </Button>
                  </Space>,
                  <Button 
                    key="bookmark"
                    type="text" 
                    icon={<StarOutlined />}
                    onClick={() => handleBookmark(post.id)}
                    style={{ color: post.isBookmarked ? '#faad14' : undefined }}
                  >
                    收藏
                  </Button>
                ]}
              >
                <List.Item.Meta
                  avatar={
                    <Avatar 
                      icon={<UserOutlined />} 
                      src={post.author.avatar}
                    />
                  }
                  title={
                    <Space>
                      <Text strong style={{ fontSize: 16 }}>{post.title}</Text>
                      <Tag color={getLevelColor(post.author.level)}>
                        L{post.author.level}
                      </Tag>
                    </Space>
                  }
                  description={
                    <Space split={<Divider type="vertical" />}>
                      <Text>{post.author.username}</Text>
                      <Text type="secondary">{post.author.title}</Text>
                      <Text type="secondary">{formatTime(post.createdAt)}</Text>
                      <Tag>{post.category}</Tag>
                    </Space>
                  }
                />
                <Paragraph ellipsis={{ rows: 2, expandable: true }}>
                  {post.content}
                </Paragraph>
                <Space>
                  {post.tags.map(tag => (
                    <Tag key={tag} color="blue">{tag}</Tag>
                  ))}
                </Space>
              </List.Item>
            )}
          />
        </div>
      )
    },
    {
      key: 'users',
      label: (
        <span>
          <TeamOutlined />
          修行者
        </span>
      ),
      children: (
        <div>
          <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
            <Col span={12}>
              <Input.Search 
                placeholder="搜索修行者..." 
                onSearch={(value) => console.log('搜索用户:', value)}
              />
            </Col>
            <Col span={6}>
              <Select placeholder="修行等级" style={{ width: '100%' }}>
                <Option value="all">全部等级</Option>
                <Option value="1-3">初学者 (L1-L3)</Option>
                <Option value="4-6">进阶者 (L4-L6)</Option>
                <Option value="7-9">导师级 (L7-L9)</Option>
              </Select>
            </Col>
          </Row>

          <Row gutter={[16, 16]}>
            {users.map(user => (
              <Col xs={24} sm={12} md={8} lg={6} key={user.id}>
                <Card
                  hoverable
                  actions={[
                    <Tooltip title="关注" key="follow">
                      <Button type="text" icon={<PlusOutlined />} />
                    </Tooltip>,
                    <Tooltip title="私信" key="message">
                      <Button type="text" icon={<MessageOutlined />} />
                    </Tooltip>
                  ]}
                >
                  <Card.Meta
                    avatar={
                      <Badge 
                        count={`L${user.level}`} 
                        style={{ backgroundColor: getLevelColor(user.level) }}
                      >
                        <Avatar 
                          size={64} 
                          icon={<UserOutlined />}
                          src={user.avatar}
                        />
                      </Badge>
                    }
                    title={user.username}
                    description={user.title}
                  />
                  <Divider />
                  <Row gutter={8}>
                    <Col span={8} style={{ textAlign: 'center' }}>
                      <div>
                        <Text strong>{user.posts}</Text>
                        <br />
                        <Text type="secondary" style={{ fontSize: 12 }}>帖子</Text>
                      </div>
                    </Col>
                    <Col span={8} style={{ textAlign: 'center' }}>
                      <div>
                        <Text strong>{user.followers}</Text>
                        <br />
                        <Text type="secondary" style={{ fontSize: 12 }}>关注者</Text>
                      </div>
                    </Col>
                    <Col span={8} style={{ textAlign: 'center' }}>
                      <div>
                        <Text strong>{user.following}</Text>
                        <br />
                        <Text type="secondary" style={{ fontSize: 12 }}>关注</Text>
                      </div>
                    </Col>
                  </Row>
                </Card>
              </Col>
            ))}
          </Row>
        </div>
      )
    },
    {
      key: 'activities',
      label: (
        <span>
          <FireOutlined />
          热门活动
        </span>
      ),
      children: (
        <div style={{ textAlign: 'center', padding: '60px 0' }}>
          <BookOutlined style={{ fontSize: 64, color: '#d9d9d9' }} />
          <Title level={4} type="secondary">活动功能开发中</Title>
          <Paragraph type="secondary">
            即将推出修行挑战、读书会、线上讲座等精彩活动
          </Paragraph>
        </div>
      )
    }
  ];

  return (
    <Layout style={{ minHeight: '100vh', background: '#f5f5f5' }}>
      <Content style={{ padding: '24px' }}>
        <div style={{ maxWidth: 1200, margin: '0 auto' }}>
          {/* 页面标题 */}
          <div style={{ marginBottom: 24 }}>
            <Title level={2}>
              <TeamOutlined style={{ marginRight: 8 }} />
              {t('community.title')}
            </Title>
            <Paragraph type="secondary">
              {t('community.subtitle')}
            </Paragraph>
          </div>

          {/* 社区统计 */}
          <Row gutter={16} style={{ marginBottom: 24 }}>
            <Col xs={24} sm={6}>
              <Card>
                <div style={{ textAlign: 'center' }}>
                  <Title level={3} style={{ margin: 0, color: '#1890ff' }}>1,234</Title>
                  <Text type="secondary">{t('community.stats.activeUsers')}</Text>
                </div>
              </Card>
            </Col>
            <Col xs={24} sm={6}>
              <Card>
                <div style={{ textAlign: 'center' }}>
                  <Title level={3} style={{ margin: 0, color: '#52c41a' }}>{posts.length}</Title>
                  <Text type="secondary">{t('community.stats.posts')}</Text>
                </div>
              </Card>
            </Col>
            <Col xs={24} sm={6}>
              <Card>
                <div style={{ textAlign: 'center' }}>
                  <Title level={3} style={{ margin: 0, color: '#faad14' }}>12,345</Title>
                  <Text type="secondary">{t('community.stats.wisdomShares')}</Text>
                </div>
              </Card>
            </Col>
            <Col xs={24} sm={6}>
              <Card>
                <div style={{ textAlign: 'center' }}>
                  <Title level={3} style={{ margin: 0, color: '#f5222d' }}>89</Title>
                  <Text type="secondary">在线用户</Text>
                </div>
              </Card>
            </Col>
          </Row>

          {/* 主要内容 */}
          <Card>
            <Tabs 
              activeKey={activeTab} 
              onChange={setActiveTab}
              items={tabItems}
            />
          </Card>
        </div>
      </Content>

      {/* 发布新帖模态框 */}
      <Modal
        title="发布新帖"
        open={showNewPost}
        onCancel={() => {
          setShowNewPost(false);
          newPostForm.resetFields();
        }}
        footer={null}
        width={600}
      >
        <Form
          form={newPostForm}
          layout="vertical"
          onFinish={handleSubmitPost}
        >
          <Form.Item
            name="title"
            label="帖子标题"
            rules={[
              { required: true, message: '请输入帖子标题' },
              { min: 5, message: '标题至少5个字符' },
              { max: 100, message: '标题不能超过100个字符' }
            ]}
          >
            <Input placeholder="请输入一个吸引人的标题..." />
          </Form.Item>

          <Form.Item
            name="category"
            label="分类"
            rules={[{ required: true, message: '请选择分类' }]}
          >
            <Select placeholder="选择帖子分类">
              <Option value="道学讨论">道学讨论</Option>
              <Option value="修行心得">修行心得</Option>
              <Option value="儒学应用">儒学应用</Option>
              <Option value="佛学智慧">佛学智慧</Option>
              <Option value="生活感悟">生活感悟</Option>
              <Option value="问题求助">问题求助</Option>
            </Select>
          </Form.Item>

          <Form.Item
            name="tags"
            label="标签"
          >
            <Select
              mode="tags"
              placeholder="添加相关标签（可选）"
              maxTagCount={5}
            >
              <Option value="道德经">道德经</Option>
              <Option value="论语">论语</Option>
              <Option value="静坐">静坐</Option>
              <Option value="冥想">冥想</Option>
              <Option value="修行">修行</Option>
              <Option value="哲学">哲学</Option>
              <Option value="智慧">智慧</Option>
            </Select>
          </Form.Item>

          <Form.Item
            name="content"
            label="内容"
            rules={[
              { required: true, message: '请输入帖子内容' },
              { min: 20, message: '内容至少20个字符' },
              { max: 5000, message: '内容不能超过5000个字符' }
            ]}
          >
            <TextArea
              rows={8}
              placeholder="分享你的想法、心得或问题..."
              showCount
              maxLength={5000}
            />
          </Form.Item>

          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={() => {
                setShowNewPost(false);
                newPostForm.resetFields();
              }}>
                取消
              </Button>
              <Button type="primary" htmlType="submit" loading={loading}>
                发布帖子
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </Layout>
  );
};

export default Community;