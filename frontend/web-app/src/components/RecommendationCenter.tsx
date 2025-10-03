import React, { useState } from 'react';
import { Card, Row, Col, Input, Button, Tabs, Typography, Space } from 'antd';
import { SearchOutlined, RobotOutlined, HeartOutlined, EyeOutlined, FireOutlined } from '@ant-design/icons';
import PersonalizedRecommendations from './PersonalizedRecommendations';
import WisdomRecommendations from './WisdomRecommendations';

const { Title, Text } = Typography;
const { TabPane } = Tabs;

const RecommendationCenter: React.FC = () => {
  const [searchWisdomId, setSearchWisdomId] = useState<string>('');

  const handleSearch = () => {
    // 搜索相似智慧的逻辑
    console.log('搜索相似智慧:', searchWisdomId);
  };

  return (
    <div style={{ padding: '24px' }}>
      <div style={{ marginBottom: '24px' }}>
        <Title level={2}>
          <RobotOutlined style={{ marginRight: '8px' }} />
          智慧推荐中心
        </Title>
        <Text type="secondary">
          基于AI算法为您推荐相关的文化智慧内容
        </Text>
      </div>

      <Tabs defaultActiveKey="personalized" size="large">
        <TabPane 
          tab={
            <span>
              <RobotOutlined />
              个性化推荐
            </span>
          } 
          key="personalized"
        >
          <PersonalizedRecommendations />
        </TabPane>

        <TabPane 
          tab={
            <span>
              <SearchOutlined />
              相似智慧
            </span>
          } 
          key="similar"
        >
          <Card>
            <Space direction="vertical" style={{ width: '100%' }}>
              <div>
                <Text strong>输入智慧ID查找相似内容：</Text>
              </div>
              <Input.Group compact>
                <Input
                  style={{ width: 'calc(100% - 80px)' }}
                  placeholder="请输入智慧ID"
                  value={searchWisdomId}
                  onChange={(e) => setSearchWisdomId(e.target.value)}
                  onPressEnter={handleSearch}
                />
                <Button 
                  type="primary" 
                  icon={<SearchOutlined />}
                  onClick={handleSearch}
                >
                  搜索
                </Button>
              </Input.Group>
              {searchWisdomId && (
                <WisdomRecommendations 
                  wisdomId={searchWisdomId}
                  title="相似智慧推荐"
                />
              )}
            </Space>
          </Card>
        </TabPane>

        <TabPane 
          tab={
            <span>
              <FireOutlined />
              热门推荐
            </span>
          } 
          key="trending"
        >
          <Card>
            <div style={{ textAlign: 'center', padding: '40px' }}>
              <FireOutlined style={{ fontSize: '48px', color: '#ff4d4f', marginBottom: '16px' }} />
              <Title level={4}>热门推荐</Title>
              <Text type="secondary">基于用户行为分析的热门智慧内容</Text>
              <div style={{ marginTop: '16px' }}>
                <Text type="secondary">功能开发中...</Text>
              </div>
            </div>
          </Card>
        </TabPane>

        <TabPane 
          tab={
            <span>
              <HeartOutlined />
              收藏推荐
            </span>
          } 
          key="favorites"
        >
          <Card>
            <div style={{ textAlign: 'center', padding: '40px' }}>
              <HeartOutlined style={{ fontSize: '48px', color: '#eb2f96', marginBottom: '16px' }} />
              <Title level={4}>基于收藏的推荐</Title>
              <Text type="secondary">根据您的收藏偏好推荐相关内容</Text>
              <div style={{ marginTop: '16px' }}>
                <Text type="secondary">功能开发中...</Text>
              </div>
            </div>
          </Card>
        </TabPane>

        <TabPane 
          tab={
            <span>
              <EyeOutlined />
              浏览历史
            </span>
          } 
          key="history"
        >
          <Card>
            <div style={{ textAlign: 'center', padding: '40px' }}>
              <EyeOutlined style={{ fontSize: '48px', color: '#1890ff', marginBottom: '16px' }} />
              <Title level={4}>基于浏览历史的推荐</Title>
              <Text type="secondary">根据您的浏览历史推荐相关内容</Text>
              <div style={{ marginTop: '16px' }}>
                <Text type="secondary">功能开发中...</Text>
              </div>
            </div>
          </Card>
        </TabPane>
      </Tabs>
    </div>
  );
};

export default RecommendationCenter;