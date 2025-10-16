import React, { useState, useEffect } from 'react';
import { Card, List, Button, Spin, Alert, Tag, Rate, Typography, Space, Empty, Select } from 'antd';
import { ReloadOutlined, EyeOutlined, HeartOutlined, BookOutlined, UserOutlined, RobotOutlined } from '@ant-design/icons';
import { apiClient } from '../services/api';
import { useAuthContext as useAuth } from '../contexts/AuthContext';

const { Title, Text, Paragraph } = Typography;
const { Option } = Select;

interface RecommendationItem {
  wisdom_id: string;
  title: string;
  author: string;
  category: string;
  school: string;
  summary: string;
  score: number;
  reason: string;
  view_count: number;
  like_count: number;
  created_at: string;
}

interface PersonalizedRecommendationsProps {
  limit?: number;
  showTitle?: boolean;
  showRefresh?: boolean;
  showAlgorithmSelector?: boolean;
  onWisdomClick?: (wisdomId: string) => void;
}

const PersonalizedRecommendations: React.FC<PersonalizedRecommendationsProps> = ({
  limit = 10,
  showTitle = true,
  showRefresh = true,
  showAlgorithmSelector = true,
  onWisdomClick
}) => {
  const { user } = useAuth();
  const [recommendations, setRecommendations] = useState<RecommendationItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [algorithm, setAlgorithm] = useState<string>('hybrid');

  const fetchPersonalizedRecommendations = async () => {
    if (!user?.id) {
      setError('请先登录以获取个性化推荐');
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const response = await apiClient.getPersonalizedRecommendations(user.id, {
        limit,
        algorithm
      });

      if (response.success && response.data) {
        setRecommendations(response.data);
      } else {
        setError(response.error || '获取个性化推荐失败');
      }
    } catch (err: any) {
      console.error('Failed to fetch personalized recommendations:', err);
      setError(err.message || '网络请求失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchPersonalizedRecommendations();
  }, [user?.id, limit, algorithm]);

  const handleWisdomClick = (item: RecommendationItem) => {
    if (onWisdomClick) {
      onWisdomClick(item.wisdom_id);
    }
  };

  const handleAlgorithmChange = (value: string) => {
    setAlgorithm(value);
  };

  const getScoreColor = (score: number) => {
    if (score >= 0.8) return '#52c41a';
    if (score >= 0.6) return '#faad14';
    return '#ff4d4f';
  };

  const getAlgorithmLabel = (alg: string) => {
    switch (alg) {
      case 'content':
        return '内容相似';
      case 'collaborative':
        return '协同过滤';
      case 'hybrid':
        return '混合推荐';
      default:
        return '混合推荐';
    }
  };

  const renderRecommendationItem = (item: RecommendationItem) => (
    <List.Item
      key={item.wisdom_id}
      className="recommendation-item"
      style={{ cursor: 'pointer', padding: '16px', borderRadius: '8px' }}
      onClick={() => handleWisdomClick(item)}
    >
      <div style={{ width: '100%' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '8px' }}>
          <Title level={5} style={{ margin: 0, flex: 1 }}>
            {item.title}
          </Title>
          <div style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
            <Rate
              disabled
              value={item.score * 5}
              allowHalf
              style={{ fontSize: '12px' }}
            />
            <Text
              style={{
                color: getScoreColor(item.score),
                fontWeight: 'bold',
                fontSize: '12px'
              }}
            >
              {(item.score * 100).toFixed(0)}%
            </Text>
          </div>
        </div>

        <Space size="small" style={{ marginBottom: '8px' }}>
          <Tag color="blue">{item.author}</Tag>
          <Tag color="green">{item.category}</Tag>
          {item.school && <Tag color="purple">{item.school}</Tag>}
        </Space>

        <Paragraph
          ellipsis={{ rows: 2, expandable: false }}
          style={{ margin: '8px 0', color: '#666' }}
        >
          {item.summary}
        </Paragraph>

        {item.reason && (
          <div style={{ marginBottom: '8px' }}>
            <Text type="secondary" style={{ fontSize: '12px' }}>
              <RobotOutlined style={{ marginRight: '4px' }} />
              推荐理由: {item.reason}
            </Text>
          </div>
        )}

        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Space size="large">
            <Space size="small">
              <EyeOutlined style={{ color: '#666' }} />
              <Text type="secondary" style={{ fontSize: '12px' }}>
                {item.view_count}
              </Text>
            </Space>
            <Space size="small">
              <HeartOutlined style={{ color: '#ff4d4f' }} />
              <Text type="secondary" style={{ fontSize: '12px' }}>
                {item.like_count}
              </Text>
            </Space>
          </Space>
          <Text type="secondary" style={{ fontSize: '12px' }}>
            {new Date(item.created_at).toLocaleDateString()}
          </Text>
        </div>
      </div>
    </List.Item>
  );

  if (!user) {
    return (
      <Card>
        <Empty
          image={<UserOutlined style={{ fontSize: '48px', color: '#d9d9d9' }} />}
          description="请先登录以获取个性化推荐"
        />
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <Alert
          message="个性化推荐加载失败"
          description={error}
          type="error"
          showIcon
          action={
            <Button size="small" onClick={fetchPersonalizedRecommendations}>
              重试
            </Button>
          }
        />
      </Card>
    );
  }

  return (
    <Card
      title={
        showTitle && (
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
              <UserOutlined />
              <span>个性化推荐</span>
            </div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
              {showAlgorithmSelector && (
                <Select
                  value={algorithm}
                  onChange={handleAlgorithmChange}
                  size="small"
                  style={{ width: 120 }}
                >
                  <Option value="hybrid">混合推荐</Option>
                  <Option value="content">内容相似</Option>
                  <Option value="collaborative">协同过滤</Option>
                </Select>
              )}
              {showRefresh && (
                <Button
                  type="text"
                  size="small"
                  icon={<ReloadOutlined />}
                  loading={loading}
                  onClick={fetchPersonalizedRecommendations}
                >
                  刷新
                </Button>
              )}
            </div>
          </div>
        )
      }
      styles={{ body: { padding: recommendations.length > 0 ? '0' : '24px' } }}
    >
      {loading ? (
        <div style={{ textAlign: 'center', padding: '40px' }}>
          <Spin size="large" />
          <div style={{ marginTop: '16px' }}>
            <Text type="secondary">正在获取个性化推荐...</Text>
          </div>
        </div>
      ) : recommendations.length > 0 ? (
        <>
          <div style={{ padding: '16px 24px 8px', borderBottom: '1px solid #f0f0f0' }}>
            <Text type="secondary" style={{ fontSize: '12px' }}>
              基于您的阅读偏好，使用 {getAlgorithmLabel(algorithm)} 算法为您推荐
            </Text>
          </div>
          <List
            dataSource={recommendations}
            renderItem={renderRecommendationItem}
            split={false}
            style={{
              maxHeight: '600px',
              overflowY: 'auto'
            }}
          />
        </>
      ) : (
        <div style={{ textAlign: 'center', padding: '40px' }}>
          <BookOutlined style={{ fontSize: '48px', color: '#d9d9d9' }} />
          <div style={{ marginTop: '16px' }}>
            <Text type="secondary">暂无个性化推荐</Text>
            <div style={{ marginTop: '8px' }}>
              <Text type="secondary" style={{ fontSize: '12px' }}>
                多阅读一些智慧内容，我们将为您提供更精准的推荐
              </Text>
            </div>
          </div>
        </div>
      )}
    </Card>
  );
};

export default PersonalizedRecommendations;