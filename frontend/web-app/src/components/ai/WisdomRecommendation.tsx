import React, { useState, useEffect } from 'react';
import { 
  Card, 
  List, 
  Button, 
  Typography, 
  Space, 
  Spin, 
  message,
  Tag,
  Progress,
  Empty
} from 'antd';
import { 
  RobotOutlined, 
  ArrowRightOutlined,
  ReloadOutlined,
  StarOutlined
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { apiClient } from '../../services/api';

const { Title, Paragraph } = Typography;

interface WisdomRecommendationProps {
  wisdomId: string;
  onRecommendationClick?: (wisdomId: string) => void;
}

interface RecommendedWisdom {
  wisdom_id: string;
  title: string;
  author: string;
  category: string;
  school: string;
  summary: string;
  relevance: number;
  reason: string;
}

interface ApiError {
  response?: {
    data?: {
      message?: string;
    };
  };
}

const WisdomRecommendation: React.FC<WisdomRecommendationProps> = ({
  wisdomId,
  onRecommendationClick
}) => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [recommendations, setRecommendations] = useState<RecommendedWisdom[]>([]);
  const [hasRecommendations, setHasRecommendations] = useState(false);

  // 获取AI智慧推荐
  const loadRecommendations = async () => {
    if (!wisdomId) {
      message.error('智慧ID不能为空');
      return;
    }

    setLoading(true);
    try {
      const response = await apiClient.get(`/cultural-wisdom/ai/${wisdomId}/recommend`);
      
      if (response.data.code === 'SUCCESS') {
        const data: RecommendedWisdom[] = response.data.data;
        setRecommendations(data);
        setHasRecommendations(true);
        message.success('AI推荐获取成功');
      } else {
        message.error(response.data.message || 'AI推荐获取失败');
      }
    } catch (error) {
      const apiError = error as ApiError;
      console.error('AI推荐请求失败:', error);
      message.error(apiError.response?.data?.message || 'AI推荐请求失败，请稍后重试');
    } finally {
      setLoading(false);
    }
  };

  // 处理推荐项点击
  const handleRecommendationClick = (recommendation: RecommendedWisdom) => {
    if (onRecommendationClick) {
      onRecommendationClick(recommendation.wisdom_id);
    } else {
      navigate(`/wisdom/${recommendation.wisdom_id}`);
    }
  };

  // 重新获取推荐
  const handleRefresh = () => {
    setRecommendations([]);
    setHasRecommendations(false);
    loadRecommendations();
  };

  // 组件挂载时自动加载推荐
  useEffect(() => {
    if (wisdomId) {
      loadRecommendations();
    }
  }, [wisdomId]);

  // 渲染推荐项
  const renderRecommendationItem = (item: RecommendedWisdom) => (
    <List.Item
      key={item.wisdom_id}
      className="hover:bg-gray-50 transition-colors duration-200 cursor-pointer rounded-lg p-3"
      onClick={() => handleRecommendationClick(item)}
    >
      <div className="w-full space-y-3">
        {/* 标题和相关度 */}
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <Title level={5} className="mb-1 text-gray-800 hover:text-blue-600 transition-colors">
              {item.title}
            </Title>
            <div className="flex items-center space-x-2 mb-2">
              <Tag color="blue">{item.category}</Tag>
              <Tag color="green">{item.school}</Tag>
              {item.author && <Tag color="orange">{item.author}</Tag>}
            </div>
          </div>
          <div className="ml-4 text-right">
            <div className="flex items-center space-x-1 mb-1">
              <StarOutlined className="text-yellow-500 text-sm" />
              <span className="font-semibold text-sm">
                {Math.round(item.relevance * 100)}%
              </span>
            </div>
            <Progress 
              percent={Math.round(item.relevance * 100)} 
              size="small" 
              strokeColor={{
                '0%': '#108ee9',
                '100%': '#87d068',
              }}
              showInfo={false}
              className="w-16"
            />
          </div>
        </div>

        {/* 摘要 */}
        <Paragraph 
          className="text-gray-600 text-sm mb-2 leading-relaxed"
          ellipsis={{ rows: 2, expandable: false }}
        >
          {item.summary}
        </Paragraph>

        {/* 推荐理由 */}
        <div className="bg-blue-50 p-2 rounded border-l-4 border-blue-400">
          <span className="text-blue-700 text-xs">
            <RobotOutlined className="mr-1" />
            推荐理由：{item.reason}
          </span>
        </div>

        {/* 操作按钮 */}
        <div className="flex justify-end">
          <Button 
            type="link" 
            size="small"
            icon={<ArrowRightOutlined />}
            className="text-blue-500 hover:text-blue-600"
          >
            查看详情
          </Button>
        </div>
      </div>
    </List.Item>
  );

  return (
    <Card 
      title={
        <Space>
          <RobotOutlined className="text-cultural-gold" />
          <span>AI智慧推荐</span>
          {hasRecommendations && (
            <Tag color="processing">{recommendations.length} 条推荐</Tag>
          )}
        </Space>
      }
      extra={
        hasRecommendations && (
          <Button 
            icon={<ReloadOutlined />} 
            onClick={handleRefresh}
            disabled={loading}
            size="small"
          >
            刷新推荐
          </Button>
        )
      }
      className="shadow-lg"
    >
      <Spin spinning={loading}>
        {!hasRecommendations && !loading && (
          <div className="text-center py-8">
            <Button 
              type="primary" 
              icon={<RobotOutlined />}
              onClick={loadRecommendations}
              size="large"
              className="bg-gradient-to-r from-blue-500 to-indigo-500 border-0 shadow-lg hover:shadow-xl transition-all duration-300"
            >
              获取AI智慧推荐
            </Button>
            <div className="mt-3 text-gray-500 text-sm">
              AI将为您推荐相关的智慧内容
            </div>
          </div>
        )}

        {hasRecommendations && recommendations.length > 0 && (
          <List
            dataSource={recommendations}
            renderItem={renderRecommendationItem}
            className="space-y-2"
          />
        )}

        {hasRecommendations && recommendations.length === 0 && (
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description="暂无相关推荐"
            className="py-8"
          >
            <Button 
              type="primary" 
              onClick={handleRefresh}
              icon={<ReloadOutlined />}
            >
              重新获取推荐
            </Button>
          </Empty>
        )}
      </Spin>
    </Card>
  );
};

export default WisdomRecommendation;