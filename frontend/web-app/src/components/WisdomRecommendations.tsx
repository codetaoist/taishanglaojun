import React, { useState, useEffect } from 'react';
import { Card, List, Button, Spin, Alert, Tag, Rate, Typography, Space, Divider } from 'antd';
import { ReloadOutlined, EyeOutlined, HeartOutlined, BookOutlined } from '@ant-design/icons';
import { apiClient } from '../services/api';

const { Title, Text, Paragraph } = Typography;

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

interface WisdomRecommendationsProps {
  wisdomId: string;
  limit?: number;
  algorithm?: 'content' | 'collaborative' | 'hybrid';
  showTitle?: boolean;
  showRefresh?: boolean;
  onWisdomClick?: (wisdomId: string) => void;
}

const WisdomRecommendations: React.FC<WisdomRecommendationsProps> = ({
  wisdomId,
  limit = 5,
  algorithm = 'hybrid',
  showTitle = true,
  showRefresh = true,
  onWisdomClick
}) => {
  const [recommendations, setRecommendations] = useState<RecommendationItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchRecommendations = async () => {
    if (!wisdomId) return;

    setLoading(true);
    setError(null);

    try {
      const response = await apiClient.getRecommendations(wisdomId, {
        limit,
        algorithm
      });

      if (response.success) {
        setRecommendations(response.data || []);
      } else {
        setError(response.message || '获取推荐失败');
      }
    } catch (err: any) {
      console.error('Failed to fetch recommendations:', err);
      setError(err.message || '网络请求失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchRecommendations();
  }, [wisdomId, limit, algorithm]);

  const handleWisdomClick = (item: RecommendationItem) => {
    if (onWisdomClick) {
      onWisdomClick(item.wisdom_id);
    }
  };

  const getScoreColor = (score: number) => {
    if (score >= 0.8) return '#52c41a';
    if (score >= 0.6) return '#faad14';
    return '#ff4d4f';
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

  if (error) {
    return (
      <Card>
        <Alert
          message="推荐加载失败"
          description={error}
          type="error"
          showIcon
          action={
            <Button size="small" onClick={fetchRecommendations}>
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
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <BookOutlined />
            <span>相关推荐</span>
            {showRefresh && (
              <Button
                type="text"
                size="small"
                icon={<ReloadOutlined />}
                loading={loading}
                onClick={fetchRecommendations}
                style={{ marginLeft: 'auto' }}
              >
                刷新
              </Button>
            )}
          </div>
        )
      }
      bodyStyle={{ padding: recommendations.length > 0 ? '0' : '24px' }}
    >
      {loading ? (
        <div style={{ textAlign: 'center', padding: '40px' }}>
          <Spin size="large" />
          <div style={{ marginTop: '16px' }}>
            <Text type="secondary">正在获取推荐...</Text>
          </div>
        </div>
      ) : recommendations.length > 0 ? (
        <List
          dataSource={recommendations}
          renderItem={renderRecommendationItem}
          split={false}
          style={{
            maxHeight: '600px',
            overflowY: 'auto'
          }}
        />
      ) : (
        <div style={{ textAlign: 'center', padding: '40px' }}>
          <BookOutlined style={{ fontSize: '48px', color: '#d9d9d9' }} />
          <div style={{ marginTop: '16px' }}>
            <Text type="secondary">暂无相关推荐</Text>
          </div>
        </div>
      )}
    </Card>
  );
};

export default WisdomRecommendations;