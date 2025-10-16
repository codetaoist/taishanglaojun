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
  , UserOutlined
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { apiClient } from '../../services/api';
import { useTranslation } from 'react-i18next';
import dashboardService, { type DashboardStats as AdminServiceStats, type SystemMetrics } from '../../services/dashboardService';

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
  const { t } = useTranslation();
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
  const [adminStats, setAdminStats] = useState<AdminServiceStats | null>(null);
  const [systemMetrics, setSystemMetrics] = useState<SystemMetrics | null>(null);
  const [healthStatus, setHealthStatus] = useState<any>(null);

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

      // 获取系统与用户统计
      const [serviceStatsRes, systemMetricsRes, healthRes] = await Promise.all([
        dashboardService.getDashboardStats(),
        dashboardService.getSystemMetrics(),
        dashboardService.getHealthStatus()
      ]);

      if (serviceStatsRes?.success) {
        setAdminStats(serviceStatsRes.data);
      }
      if (systemMetricsRes?.success) {
        setSystemMetrics(systemMetricsRes.data);
      }
      if (healthRes?.success) {
        setHealthStatus(healthRes.data);
      }
    } catch {
      console.error(t('adminDashboard.error.loadFailed'));
    }
  };

  useEffect(() => {
    loadDashboardData();
  }, []);

  // 快捷操作按钮
  const quickActions = [
    {
      title: t('adminDashboard.quickActions.addWisdom'),
      icon: <PlusOutlined />,
      color: '#1890ff',
      onClick: () => navigate('/admin/wisdom/create')
    },
    {
      title: t('adminDashboard.quickActions.manageCategories'),
      icon: <FolderOutlined />,
      color: '#52c41a',
      onClick: () => navigate('/admin/categories')
    },
    {
      title: t('adminDashboard.quickActions.manageTags'),
      icon: <TagOutlined />,
      color: '#faad14',
      onClick: () => navigate('/admin/tags')
    },
    {
      title: t('adminDashboard.quickActions.systemSettings'),
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
              {t('adminDashboard.title')}
            </Title>
            <Paragraph className="text-gray-600 mb-0">
              {t('adminDashboard.subtitle')}
            </Paragraph>
          </div>
          <Button 
            type="primary" 
            icon={<PlusOutlined />}
            size="large"
            onClick={() => navigate('/admin/wisdom/create')}
          >
            {t('adminDashboard.quickActions.addWisdom')}
          </Button>
        </div>
      </Card>

      {/* 用户与系统概览 */}
      <Row gutter={[16, 16]}>
        <Col xs={24} md={12}>
          <Card title={t('adminDashboard.sections.systemOverview')}>
            <Row gutter={[16, 16]}>
              <Col xs={12}>
                <Statistic
                  title={t('adminDashboard.stats.totalUsers')}
                  value={adminStats?.totalUsers ?? 0}
                  prefix={<UserOutlined className="text-blue-500" />}
                />
              </Col>
              <Col xs={12}>
                <Statistic
                  title={t('adminDashboard.stats.activeUsers')}
                  value={adminStats?.activeUsers ?? 0}
                  valueStyle={{ color: '#52c41a' }}
                />
              </Col>
            </Row>
            <div className="mt-4">
              <Statistic
                title={t('adminDashboard.stats.systemHealth')}
                value={adminStats?.systemHealth ?? 0}
                suffix="%"
                valueStyle={{ color: (adminStats?.systemHealth ?? 0) > 90 ? '#52c41a' : '#faad14' }}
              />
            </div>
          </Card>
        </Col>
        <Col xs={24} md={12}>
          <Card title={t('adminDashboard.sections.systemMetrics')}>
            <Space direction="vertical" size="middle" style={{ width: '100%' }}>
              <div>
                <div className="flex items-center justify-between mb-2">
                  <span className="font-medium">{t('adminDashboard.metrics.cpu')}</span>
                  <Tag color="blue">{systemMetrics?.cpu?.usage ?? adminStats?.cpuUsage ?? 0}%</Tag>
                </div>
                <Progress percent={Math.round(systemMetrics?.cpu?.usage ?? adminStats?.cpuUsage ?? 0)} />
              </div>
              <div>
                <div className="flex items-center justify-between mb-2">
                  <span className="font-medium">{t('adminDashboard.metrics.memory')}</span>
                  <Tag color="green">{systemMetrics?.memory?.usage ?? adminStats?.memoryUsage ?? 0}%</Tag>
                </div>
                <Progress percent={Math.round(systemMetrics?.memory?.usage ?? adminStats?.memoryUsage ?? 0)} status={(systemMetrics?.memory?.usage ?? adminStats?.memoryUsage ?? 0) > 85 ? 'exception' : undefined} />
              </div>
              <div>
                <div className="flex items-center justify-between mb-2">
                  <span className="font-medium">{t('adminDashboard.metrics.disk')}</span>
                  <Tag color="orange">{systemMetrics?.disk?.usage ?? adminStats?.diskUsage ?? 0}%</Tag>
                </div>
                <Progress percent={Math.round(systemMetrics?.disk?.usage ?? adminStats?.diskUsage ?? 0)} />
              </div>
              <div>
                <div className="flex items-center justify-between mb-2">
                  <span className="font-medium">{t('adminDashboard.metrics.networkLatency')}</span>
                  <Tag color="purple">{systemMetrics?.network?.latency ?? adminStats?.networkLatency ?? 0}ms</Tag>
                </div>
                <Progress percent={Math.min(100, Math.round(((systemMetrics?.network?.latency ?? adminStats?.networkLatency ?? 0) / 100) * 100))} />
              </div>
            </Space>
          </Card>
        </Col>
      </Row>

      {/* 统计卡片 */}
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title={t('adminDashboard.stats.totalWisdom')}
              value={stats.totalWisdom}
              prefix={<BookOutlined className="text-blue-500" />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title={t('adminDashboard.stats.totalViews')}
              value={stats.totalViews}
              prefix={<EyeOutlined className="text-green-500" />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title={t('adminDashboard.stats.totalLikes')}
              value={stats.totalLikes}
              prefix={<HeartOutlined className="text-red-500" />}
              valueStyle={{ color: '#f5222d' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title={t('adminDashboard.stats.totalCategories')}
              value={stats.totalCategories}
              prefix={<FolderOutlined className="text-orange-500" />}
              valueStyle={{ color: '#fa8c16' }}
            />
          </Card>
        </Col>
      </Row>

      {/* 快捷操作 */}
      <Card title={t('adminDashboard.actions.quick')}>
        <Row gutter={[16, 16]}>
          {quickActions.map((action, index) => (
            <Col xs={12} sm={8} md={6} key={index}>
              <Card
                hoverable
                className="text-center cursor-pointer"
                onClick={action.onClick}
                styles={{ body: { padding: '24px 16px' } }}
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
            title={t('adminDashboard.lists.recentAdded')} 
            extra={
              <Button 
                type="link" 
                onClick={() => navigate('/admin/wisdom')}
              >
                {t('adminDashboard.lists.viewAll')}
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
                      {t('adminDashboard.lists.edit')}
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
            title={t('adminDashboard.lists.popularContent')} 
            extra={
              <Button 
                type="link"
                onClick={() => navigate('/admin/wisdom?sort=views')}
              >
                {t('adminDashboard.lists.viewAll')}
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
      <Card title={t('adminDashboard.sections.categoryStats')}>
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