import React, { useState } from 'react';
import { Card, Button, Typography, Space, Divider, message, Spin } from 'antd';
import { ApiOutlined, SearchOutlined, MessageOutlined, RobotOutlined, BookOutlined } from '@ant-design/icons';
import { apiClient } from '../services/api';

const { Title, Paragraph } = Typography;

interface TestResults {
  health?: unknown;
  wisdomList?: unknown;
  wisdomDetail?: unknown;
  search?: unknown;
  aiProviders?: unknown;
  aiChat?: unknown;
}

const ApiTest: React.FC = () => {
  const [loading, setLoading] = useState({
    health: false,
    wisdomList: false,
    wisdomDetail: false,
    search: false,
    aiProviders: false,
    aiChat: false
  });

  const [results, setResults] = useState<TestResults>({});

  const testHealthCheck = async () => {
    setLoading(prev => ({ ...prev, health: true }));
    try {
      const response = await fetch('http://localhost:8080/health');
      const data = await response.json();
      setResults((prev) => ({ ...prev, health: data }));
      message.success('健康检查成功');
    } catch (error) {
      message.error('健康检查失败');
      console.error(error);
    }
    setLoading((prev) => ({ ...prev, health: false }));
  };

  const testWisdomList = async () => {
    setLoading(prev => ({ ...prev, wisdomList: true }));
    try {
      const response = await apiClient.getWisdomList({ page: 1, pageSize: 10 });
      setResults((prev) => ({
        ...prev,
        wisdomList: {
          success: response.success,
          data: response.data,
          timestamp: new Date().toISOString()
        }
      }));
      if (response.success) {
        message.success('智慧列表测试成功');
      } else {
        message.error('智慧列表测试失败');
      }
    } catch (error) {
       console.error('Wisdom list test error:', error);
       setResults((prev) => ({
         ...prev,
         wisdomList: {
           success: false,
           error: error instanceof Error ? error.message : '未知错误',
           timestamp: new Date().toISOString()
         }
       }));
       message.error('智慧列表测试失败');
     }
     setLoading((prev) => ({ ...prev, wisdomList: false }));
  };

  const testWisdomDetail = async () => {
    setLoading(prev => ({ ...prev, wisdomDetail: true }));
    try {
      const response = await apiClient.getWisdomById('1');
      setResults((prev: any) => ({
        ...prev,
        wisdomDetail: {
          success: response.success,
          data: response.data,
          timestamp: new Date().toISOString()
        }
      }));
      if (response.success) {
        message.success('智慧详情测试成功');
      } else {
        message.error('智慧详情测试失败');
      }
    } catch (error) {
       console.error('Wisdom detail test error:', error);
       setResults((prev) => ({
         ...prev,
         wisdomDetail: {
           success: false,
           error: error instanceof Error ? error.message : '未知错误',
           timestamp: new Date().toISOString()
         }
       }));
       message.error('智慧详情测试失败');
     }
     setLoading((prev) => ({ ...prev, wisdomDetail: false }));
  };

  const testAIProviders = async () => {
    setLoading(prev => ({ ...prev, aiProviders: true }));
    try {
      const response = await apiClient.getAIProviders();
      setResults((prev: any) => ({
        ...prev,
        aiProviders: {
          success: response.success,
          data: response.data,
          timestamp: new Date().toISOString()
        }
      }));
      if (response.success) {
        message.success('AI提供商测试成功');
      } else {
        message.error('AI提供商测试失败');
      }
    } catch (error) {
       console.error('AI Providers test error:', error);
       setResults((prev) => ({
         ...prev,
         aiProviders: {
           success: false,
           error: error instanceof Error ? error.message : '未知错误',
           timestamp: new Date().toISOString()
         }
       }));
       message.error('AI提供商测试失败');
     }
     setLoading((prev) => ({ ...prev, aiProviders: false }));
  };

  const testAIChat = async () => {
    setLoading(prev => ({ ...prev, aiChat: true }));
    try {
      const response = await apiClient.sendChatMessage('你好，请介绍一下太上老君的智慧');
      setResults((prev: any) => ({
        ...prev,
        aiChat: {
          success: response.success,
          data: response.data,
          timestamp: new Date().toISOString()
        }
      }));
      if (response.success) {
        message.success('AI聊天测试成功');
      } else {
        message.error('AI聊天测试失败');
      }
    } catch (error) {
       console.error('AI Chat test error:', error);
       setResults((prev) => ({
         ...prev,
         aiChat: {
           success: false,
           error: error instanceof Error ? error.message : '未知错误',
           timestamp: new Date().toISOString()
         }
       }));
       message.error('AI聊天测试失败');
     }
     setLoading((prev) => ({ ...prev, aiChat: false }));
  };

  return (
    <div className="p-6 space-y-6">
      <Card>
        <Title level={2} className="flex items-center gap-2">
          <ApiOutlined className="text-primary-500" />
          前后端集成测试
        </Title>
        <Paragraph>
          测试前端应用与后端API服务的连接状态和功能集成
        </Paragraph>
      </Card>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* 健康检查测试 */}
        <Card title="健康检查测试" extra={<ApiOutlined />}>
          <Space direction="vertical" className="w-full">
            <Button 
              type="primary" 
              onClick={testHealthCheck}
              loading={loading.health}
              className="w-full"
            >
              测试健康检查 API
            </Button>
            {results.health && (
              <div className="bg-gray-50 p-3 rounded max-h-32 overflow-y-auto">
                <pre className="text-sm">
                  {JSON.stringify(results.health, null, 2)}
                </pre>
              </div>
            )}
          </Space>
        </Card>

        {/* 智慧列表测试 */}
        <Card title="智慧列表测试" extra={<BookOutlined />}>
          <Space direction="vertical" className="w-full">
            <Button 
              type="primary" 
              onClick={testWisdomList}
              loading={loading.wisdomList}
              className="w-full"
            >
              测试智慧列表 API
            </Button>
            {results.wisdomList && (
              <div className="bg-gray-50 p-3 rounded max-h-32 overflow-y-auto">
                <pre className="text-sm">
                  {JSON.stringify(results.wisdomList, null, 2)}
                </pre>
              </div>
            )}
          </Space>
        </Card>

        {/* 智慧详情测试 */}
        <Card title="智慧详情测试" extra={<SearchOutlined />}>
          <Space direction="vertical" className="w-full">
            <Button 
              type="primary" 
              onClick={testWisdomDetail}
              loading={loading.wisdomDetail}
              className="w-full"
            >
              测试智慧详情 API
            </Button>
            {results.wisdomDetail && (
              <div className="bg-gray-50 p-3 rounded max-h-32 overflow-y-auto">
                <pre className="text-sm">
                  {JSON.stringify(results.wisdomDetail, null, 2)}
                </pre>
              </div>
            )}
          </Space>
        </Card>

        {/* AI提供商测试 */}
        <Card title="AI提供商测试" extra={<RobotOutlined />}>
          <Space direction="vertical" className="w-full">
            <Button 
              type="primary" 
              onClick={testAIProviders}
              loading={loading.aiProviders}
              className="w-full"
            >
              测试AI提供商 API
            </Button>
            {results.aiProviders && (
              <div className="bg-gray-50 p-3 rounded max-h-32 overflow-y-auto">
                <pre className="text-sm">
                  {JSON.stringify(results.aiProviders, null, 2)}
                </pre>
              </div>
            )}
          </Space>
        </Card>

        {/* AI聊天测试 */}
        <Card title="AI聊天测试" extra={<MessageOutlined />}>
          <Space direction="vertical" className="w-full">
            <Button 
              type="primary" 
              onClick={testAIChat}
              loading={loading.aiChat}
              className="w-full"
            >
              测试AI聊天 API
            </Button>
            {results.aiChat && (
              <div className="bg-gray-50 p-3 rounded max-h-32 overflow-y-auto">
                <pre className="text-sm">
                  {JSON.stringify(results.aiChat, null, 2)}
                </pre>
              </div>
            )}
          </Space>
        </Card>
      </div>

      <Divider />

      {/* 测试结果汇总 */}
      <Card title="测试结果汇总">
        <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
          <div className="text-center">
            <div className="text-lg font-semibold">健康检查</div>
            <div className={`text-sm ${results.health ? 'text-green-600' : 'text-gray-400'}`}>
              {results.health ? '✓ 通过' : '- 未测试'}
            </div>
          </div>
          <div className="text-center">
            <div className="text-lg font-semibold">智慧列表</div>
            <div className={`text-sm ${results.wisdomList ? 'text-green-600' : 'text-gray-400'}`}>
              {results.wisdomList ? '✓ 通过' : '- 未测试'}
            </div>
          </div>
          <div className="text-center">
            <div className="text-lg font-semibold">智慧详情</div>
            <div className={`text-sm ${results.wisdomDetail ? 'text-green-600' : 'text-gray-400'}`}>
              {results.wisdomDetail ? '✓ 通过' : '- 未测试'}
            </div>
          </div>
          <div className="text-center">
            <div className="text-lg font-semibold">AI提供商</div>
            <div className={`text-sm ${results.aiProviders ? 'text-green-600' : 'text-gray-400'}`}>
              {results.aiProviders ? '✓ 通过' : '- 未测试'}
            </div>
          </div>
          <div className="text-center">
            <div className="text-lg font-semibold">AI聊天</div>
            <div className={`text-sm ${results.aiChat ? 'text-green-600' : 'text-gray-400'}`}>
              {results.aiChat ? '✓ 通过' : '- 未测试'}
            </div>
          </div>
        </div>
      </Card>

      {Object.values(loading).some(Boolean) && (
        <div className="text-center">
          <Spin size="large" />
          <div className="mt-2">
            <span className="text-gray-600">正在测试API连接...</span>
          </div>
        </div>
      )}
    </div>
  );
};

export default ApiTest;