import React, { useState } from 'react';
import { Card, Row, Col, Tabs, Typography, Space, Button, Input, Select, Divider } from 'antd';
import { 
  UserOutlined, 
  BookOutlined, 
  RobotOutlined, 
  SearchOutlined,
  FireOutlined,
  StarOutlined,
  ClockCircleOutlined
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import behaviorService from '../services/behaviorService';
import PersonalizedRecommendations from '../components/PersonalizedRecommendations';
import WisdomRecommendations from '../components/WisdomRecommendations';

const { Title, Text } = Typography;
const { Search } = Input;
const { Option } = Select;

const RecommendationCenter: React.FC = () => {
  const navigate = useNavigate();
  const [searchWisdomId, setSearchWisdomId] = useState<string>('');
  const [activeTab, setActiveTab] = useState<string>('personalized');

  const handleWisdomClick = async (wisdomId: string) => {
    // 记录点击行为
    try {
      await behaviorService.recordClick('wisdom', wisdomId);
    } catch (error) {
      console.warn('记录点击行为失败:', error);
    }
    
    navigate(`/wisdom/${wisdomId}`);
  };

  const handleSearchRecommendations = async (wisdomId: string) => {
    if (wisdomId.trim()) {
      // 记录搜索行为
      try {
        await behaviorService.recordSearch(`相似推荐:${wisdomId}`, 0);
      } catch (error) {
        console.warn('记录搜索行为失败:', error);
      }
      
      setSearchWisdomId(wisdomId.trim());
      setActiveTab('similar');
    }
  };

  const tabItems = [
    {
      key: 'personalized',
      label: (
        <Space>
          <UserOutlined />
          个性化推荐
        </Space>
      ),
      children: (
        <PersonalizedRecommendations
          limit={20}
          onWisdomClick={handleWisdomClick}
        />
      ),
    },
    {
      key: 'similar',
      label: (
        <Space>
          <BookOutlined />
          相似推荐
        </Space>
      ),
      children: (
        <div>
          <Card size="small" style={{ marginBottom: '16px' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
              <Text strong>输入智慧ID获取相似推荐:</Text>
              <Search
                placeholder="请输入智慧ID"
                allowClear
                enterButton="获取推荐"
                size="middle"
                style={{ flex: 1, maxWidth: '400px' }}
                onSearch={handleSearchRecommendations}
              />
            </div>
          </Card>
          {searchWisdomId && (
            <WisdomRecommendations
              wisdomId={searchWisdomId}
              limit={15}
              algorithm="content"
              onWisdomClick={handleWisdomClick}
            />
          )}
          {!searchWisdomId && (
            <Card>
              <div style={{ textAlign: 'center', padding: '40px' }}>
                <SearchOutlined style={{ fontSize: '48px', color: '#d9d9d9' }} />
                <div style={{ marginTop: '16px' }}>
                  <Text type="secondary">请输入智慧ID以获取相似推荐</Text>
                </div>
              </div>
            </Card>
          )}
        </div>
      ),
    },
    {
      key: 'trending',
      label: (
        <Space>
          <FireOutlined />
          热门推荐
        </Space>
      ),
      children: (
        <Card>
          <div style={{ textAlign: 'center', padding: '40px' }}>
            <FireOutlined style={{ fontSize: '48px', color: '#d9d9d9' }} />
            <div style={{ marginTop: '16px' }}>
              <Text type="secondary">热门推荐功能开发中...</Text>
            </div>
          </div>
        </Card>
      ),
    },
    {
      key: 'favorites',
      label: (
        <Space>
          <StarOutlined />
          收藏推荐
        </Space>
      ),
      children: (
        <Card>
          <div style={{ textAlign: 'center', padding: '40px' }}>
            <StarOutlined style={{ fontSize: '48px', color: '#d9d9d9' }} />
            <div style={{ marginTop: '16px' }}>
              <Text type="secondary">基于收藏的推荐功能开发中...</Text>
            </div>
          </div>
        </Card>
      ),
    },
    {
      key: 'recent',
      label: (
        <Space>
          <ClockCircleOutlined />
          最近浏览
        </Space>
      ),
      children: (
        <Card>
          <div style={{ textAlign: 'center', padding: '40px' }}>
            <ClockCircleOutlined style={{ fontSize: '48px', color: '#d9d9d9' }} />
            <div style={{ marginTop: '16px' }}>
              <Text type="secondary">基于浏览历史的推荐功能开发中...</Text>
            </div>
          </div>
        </Card>
      ),
    },
  ];

  return (
    <div style={{ padding: '24px', maxWidth: '1200px', margin: '0 auto' }}>
      {/* 页面标题 */}
      <div style={{ marginBottom: '24px' }}>
        <Title level={2} style={{ margin: 0, display: 'flex', alignItems: 'center', gap: '12px' }}>
          <RobotOutlined style={{ color: '#1890ff' }} />
          智慧推荐中心
        </Title>
        <Text type="secondary" style={{ fontSize: '16px', marginTop: '8px', display: 'block' }}>
          基于AI算法为您推荐相关的文化智慧内容
        </Text>
      </div>

      {/* 功能介绍 */}
      <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
        <Col xs={24} sm={12} md={6}>
          <Card size="small" style={{ textAlign: 'center', height: '100%' }}>
            <UserOutlined style={{ fontSize: '24px', color: '#1890ff', marginBottom: '8px' }} />
            <div>
              <Text strong>个性化推荐</Text>
              <div style={{ marginTop: '4px' }}>
                <Text type="secondary" style={{ fontSize: '12px' }}>
                  基于您的阅读偏好
                </Text>
              </div>
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card size="small" style={{ textAlign: 'center', height: '100%' }}>
            <BookOutlined style={{ fontSize: '24px', color: '#52c41a', marginBottom: '8px' }} />
            <div>
              <Text strong>相似推荐</Text>
              <div style={{ marginTop: '4px' }}>
                <Text type="secondary" style={{ fontSize: '12px' }}>
                  基于内容相似度
                </Text>
              </div>
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card size="small" style={{ textAlign: 'center', height: '100%' }}>
            <FireOutlined style={{ fontSize: '24px', color: '#fa541c', marginBottom: '8px' }} />
            <div>
              <Text strong>热门推荐</Text>
              <div style={{ marginTop: '4px' }}>
                <Text type="secondary" style={{ fontSize: '12px' }}>
                  基于热度排行
                </Text>
              </div>
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card size="small" style={{ textAlign: 'center', height: '100%' }}>
            <RobotOutlined style={{ fontSize: '24px', color: '#722ed1', marginBottom: '8px' }} />
            <div>
              <Text strong>AI智能</Text>
              <div style={{ marginTop: '4px' }}>
                <Text type="secondary" style={{ fontSize: '12px' }}>
                  多算法融合
                </Text>
              </div>
            </div>
          </Card>
        </Col>
      </Row>

      <Divider />

      {/* 推荐内容 */}
      <Tabs
        activeKey={activeTab}
        onChange={setActiveTab}
        items={tabItems}
        size="large"
        style={{ minHeight: '600px' }}
      />
    </div>
  );
};

export default RecommendationCenter;