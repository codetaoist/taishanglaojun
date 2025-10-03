import React, { useState, useEffect } from 'react';
import {
  Card,
  Row,
  Col,
  Statistic,
  Typography,
  Space,
  Button,
  Tag,
  Progress,
  List,
  Avatar
} from 'antd';
import {
  BookOutlined,
  EyeOutlined,
  HeartOutlined,
  TagOutlined,
  FolderOutlined,
  PlusOutlined,
  EditOutlined,
  SettingOutlined
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { apiClient } from '../../services/api';

const { Title, Paragraph } = Typography;

interface DashboardStats {
  totalUsers: number;
  totalWisdom: number;
  totalViews: number;
  totalLikes: number;
  totalCategories: number;
  totalTags: number;
  recentWisdom: {
    id: string;
    title: string;
    author: string;
    createdAt: string;
    views: number;
  }[];
  popularWisdom: {
    id: string;
    title: string;
    views: number;
    likes: number;
  }[];
  categoryStats: {
    name: string;
    count: number;
    percentage: number;
  }[];
}

const AdminDashboard: React.FC = () => {
  const navigate = useNavigate();
  const [stats, setStats] = useState<DashboardStats>({
    totalWisdom: 0,
    totalViews: 0,
    totalLikes: 0,
    totalCategories: 0,
    totalTags: 0,
    recentWisdom: [],
    popularWisdom: [],
    categoryStats: []
  });

  // 加载仪表板数据
  const loadDashboardData = async () => {
    try {
      // 获取统计数据
      const statsResponse = await apiClient.getWisdomStats();
      
      // 获取最近的智慧内容
      const recentResponse = await apiClient.getWisdomList({
        page: 1,
        pageSize: 5,
        sortBy: 'created_at',
        sortOrder: 'desc'
      });

      // 获取热门智慧内容
      const popularResponse = await apiClient.getWisdomList({
        page: 1,
        pageSize: 5,
        sortBy: 'views',
        sortOrder: 'desc'
      });

      // 获取分类统计
      const categoriesResponse = await apiClient.getCategories();

      if (statsResponse.success) {
        setStats(prev => ({
          ...prev,
          totalWisdom: statsResponse.data?.totalWisdom || 0,
          totalViews: statsResponse.data?.totalViews || 0,
          totalLikes: statsResponse.data?.totalLikes || 0,
          totalCategories: statsResponse.data?.totalCategories || 0,
          totalTags: statsResponse.data?.totalTags || 0
        }));
      }

      if (recentResponse.success) {
        setStats(prev => ({
          ...prev,
          recentWisdom: recentResponse.data?.items || []
        }));
      }

      if (popularResponse.success) {
        setStats(prev => ({
          ...prev,
          popularWisdom: popularResponse.data?.items || []
        }));
      }

      if (categoriesResponse.success) {
        setStats(prev => ({
          ...prev,
          categoryStats: categoriesResponse.data || []
        }));
      }
    } catch {
      console.error('加载仪表板数据失败');
    }
  };

  useEffect(() => {
    loadDashboardData();
  }, []);

  // 快捷操作按钮
  const quickActions = [
    {
      title: '添加智慧',
      icon: <PlusOutlined />,
      color: '#1890ff',
      onClick: () => navigate('/admin/wisdom/create')
    },
    {
      title: '管理分类',
      icon: <FolderOutlined />,
      color: '#52c41a',
      onClick: () => navigate('/admin/categories')
    },
    {
      title: '管理标签',
      icon: <TagOutlined />,
      color: '#faad14',
      onClick: () => navigate('/admin/tags')
    },
    {
      title: '系统设置',
      icon: <SettingOutlined />,
      color: '#722ed1',
      onClick: () => navigate('/admin/settings')
    }
  ];

  return (
    <div className="space-y-6">
      {/* 页面标题 */}
      <Card className="bg-gradient-to-r from-blue-50 to-indigo-50 border-blue-200">
        <div className="flex items-center justify-between">
          <div>
            <Title level={2} className="mb-2">
              管理员仪表板
            </Title>
            <Paragraph className="text-gray-600 mb-0">
              欢迎回来！这里是文化智慧内容管理中心
            </Paragraph>
          </div>
          <Button 
            type="primary" 
            icon={<PlusOutlined />}
            size="large"
            onClick={() => navigate('/admin/wisdom/create')}
          >
            添加智慧
          </Button>
        </div>
      </Card>

      {/* 统计卡片 */}
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="智慧总数"
              value={stats.totalWisdom}
              prefix={<BookOutlined className="text-blue-500" />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="总阅读量"
              value={stats.totalViews}
              prefix={<EyeOutlined className="text-green-500" />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="总点赞数"
              value={stats.totalLikes}
              prefix={<HeartOutlined className="text-red-500" />}
              valueStyle={{ color: '#f5222d' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="分类数量"
              value={stats.totalCategories}
              prefix={<FolderOutlined className="text-orange-500" />}
              valueStyle={{ color: '#fa8c16' }}
            />
          </Card>
        </Col>
      </Row>

      {/* 快捷操作 */}
      <Card title="快捷操作">
        <Row gutter={[16, 16]}>
          {quickActions.map((action, index) => (
            <Col xs={12} sm={8} md={6} key={index}>
              <Card
                hoverable
                className="text-center cursor-pointer"
                onClick={action.onClick}
                bodyStyle={{ padding: '24px 16px' }}
              >
                <div 
                  className="text-3xl mb-3"
                  style={{ color: action.color }}
                >
                  {action.icon}
                </div>
                <div className="font-medium">{action.title}</div>
              </Card>
            </Col>
          ))}
        </Row>
      </Card>

      <Row gutter={[16, 16]}>
        {/* 最近添加的智慧 */}
        <Col xs={24} lg={12}>
          <Card 
            title="最近添加" 
            extra={
              <Button 
                type="link" 
                onClick={() => navigate('/admin/wisdom')}
              >
                查看全部
              </Button>
            }
          >
            <List
              dataSource={stats.recentWisdom}
              renderItem={(item: { id: string; title: string; category: string; created_at: string }) => (
                <List.Item
                  actions={[
                    <Button 
                      type="link" 
                      icon={<EditOutlined />}
                      onClick={() => navigate(`/admin/wisdom/${item.id}/edit`)}
                    >
                      编辑
                    </Button>
                  ]}
                >
                  <List.Item.Meta
                    avatar={<Avatar icon={<BookOutlined />} />}
                    title={
                      <a onClick={() => navigate(`/wisdom/${item.id}`)}>
                        {item.title}
                      </a>
                    }
                    description={
                      <Space>
                        <Tag color="blue">{item.category}</Tag>
                        <span className="text-gray-500">
                          {new Date(item.created_at).toLocaleDateString()}
                        </span>
                      </Space>
                    }
                  />
                </List.Item>
              )}
            />
          </Card>
        </Col>

        {/* 热门智慧 */}
        <Col xs={24} lg={12}>
          <Card 
            title="热门内容" 
            extra={
              <Button 
                type="link"
                onClick={() => navigate('/admin/wisdom?sort=views')}
              >
                查看全部
              </Button>
            }
          >
            <List
              dataSource={stats.popularWisdom}
              renderItem={(item: { id: string; title: string; views: number; likes: number }, index: number) => (
                <List.Item>
                  <List.Item.Meta
                    avatar={
                      <Avatar 
                        style={{ 
                          backgroundColor: index < 3 ? '#f56a00' : '#87d068' 
                        }}
                      >
                        {index + 1}
                      </Avatar>
                    }
                    title={
                      <a onClick={() => navigate(`/wisdom/${item.id}`)}>
                        {item.title}
                      </a>
                    }
                    description={
                      <Space>
                        <span>
                          <EyeOutlined /> {item.views || 0}
                        </span>
                        <span>
                          <HeartOutlined /> {item.likes || 0}
                        </span>
                      </Space>
                    }
                  />
                </List.Item>
              )}
            />
          </Card>
        </Col>
      </Row>

      {/* 分类统计 */}
      <Card title="分类统计">
        <Row gutter={[16, 16]}>
          {stats.categoryStats.slice(0, 6).map((category: { id: string; name: string; wisdom_count: number }) => (
            <Col xs={24} sm={12} md={8} key={category.id}>
              <Card size="small">
                <div className="flex items-center justify-between mb-2">
                  <span className="font-medium">{category.name}</span>
                  <Tag color="blue">{category.wisdom_count || 0}</Tag>
                </div>
                <Progress 
                  percent={
                    stats.totalWisdom > 0 
                      ? Math.round((category.wisdom_count || 0) / stats.totalWisdom * 100)
                      : 0
                  }
                  size="small"
                  strokeColor={{
                    '0%': '#108ee9',
                    '100%': '#87d068',
                  }}
                />
              </Card>
            </Col>
          ))}
        </Row>
      </Card>
    </div>
  );
};

export default AdminDashboard;