import React, { useState } from 'react';
import { Card, Input, Button, Typography, Space, Spin, Alert, Tag, Divider, List, Avatar } from 'antd';
import { 
  SendOutlined, 
  RobotOutlined, 
  BookOutlined, 
  BulbOutlined,
  StarOutlined,
  QuestionCircleOutlined
} from '@ant-design/icons';
import { apiClient } from '../../services/api';

const { Title, Paragraph, Text } = Typography;
const { TextArea } = Input;

interface WisdomReference {
  wisdom_id: string;
  title: string;
  author: string;
  school: string;
  excerpt: string;
  relevance: number;
}

interface QAResponse {
  question: string;
  answer: string;
  related_wisdoms: WisdomReference[];
  sources: string[];
  confidence: number;
  keywords: string[];
  category: string;
}

const IntelligentQA: React.FC = () => {
  const [question, setQuestion] = useState('');
  const [loading, setLoading] = useState(false);
  const [response, setResponse] = useState<QAResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async () => {
    if (!question.trim()) return;

    setLoading(true);
    setError(null);
    
    try {
      const result = await apiClient.post('/api/v1/cultural-wisdom/ai/qa', {
        question: question.trim()
      });
      
      setResponse(result.data);
    } catch (err: any) {
      setError(err.response?.data?.message || '智能问答失败，请稍后重试');
    } finally {
      setLoading(false);
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
      handleSubmit();
    }
  };

  const exampleQuestions = [
    '什么是道德经中的"无为而治"？',
    '孔子的"仁"思想在现代社会有什么意义？',
    '佛教的"空"与道教的"无"有什么区别？',
    '如何理解《易经》中的阴阳思想？',
    '儒家的修身齐家治国平天下如何实践？'
  ];

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      {/* 标题区域 */}
      <Card className="bg-gradient-to-r from-blue-50 to-purple-50 border-blue-200">
        <div className="text-center space-y-4">
          <div className="w-16 h-16 mx-auto bg-gradient-to-r from-blue-500 to-purple-500 rounded-2xl flex items-center justify-center shadow-lg">
            <RobotOutlined className="text-3xl text-white" />
          </div>
          <Title level={2} className="mb-2">
            AI智慧问答
          </Title>
          <Paragraph className="text-gray-600 text-lg">
            向AI导师提问，获得专业的中华传统文化解答
          </Paragraph>
        </div>
      </Card>

      {/* 问题输入区域 */}
      <Card title={
        <Space>
          <QuestionCircleOutlined className="text-blue-500" />
          <span>提出您的问题</span>
        </Space>
      }>
        <div className="space-y-4">
          <TextArea
            value={question}
            onChange={(e) => setQuestion(e.target.value)}
            onKeyDown={handleKeyPress}
            placeholder="请输入您关于中华传统文化的问题..."
            rows={4}
            maxLength={500}
            showCount
            className="text-base"
          />
          
          <div className="flex justify-between items-center">
            <Text type="secondary" className="text-sm">
              按 Ctrl+Enter 快速提交
            </Text>
            <Button
              type="primary"
              icon={<SendOutlined />}
              onClick={handleSubmit}
              loading={loading}
              disabled={!question.trim()}
              size="large"
              className="px-8"
            >
              提问
            </Button>
          </div>
        </div>
      </Card>

      {/* 示例问题 */}
      {!response && (
        <Card title="示例问题" className="bg-gray-50">
          <div className="space-y-3">
            {exampleQuestions.map((q, index) => (
              <Button
                key={index}
                type="text"
                onClick={() => setQuestion(q)}
                className="text-left h-auto p-3 w-full border border-gray-200 hover:border-blue-300 hover:bg-blue-50 rounded-lg"
              >
                <BulbOutlined className="text-yellow-500 mr-2" />
                {q}
              </Button>
            ))}
          </div>
        </Card>
      )}

      {/* 错误提示 */}
      {error && (
        <Alert
          message="问答失败"
          description={error}
          type="error"
          showIcon
          closable
          onClose={() => setError(null)}
        />
      )}

      {/* 加载状态 */}
      {loading && (
        <Card>
          <div className="text-center py-12">
            <Spin size="large" />
            <div className="mt-4 text-gray-600">AI正在思考中，请稍候...</div>
          </div>
        </Card>
      )}

      {/* 回答结果 */}
      {response && (
        <div className="space-y-6">
          {/* 问题回显 */}
          <Card className="qa-question-card">
            <div className="flex items-start space-x-3">
              <Avatar icon={<QuestionCircleOutlined />} className="bg-blue-500" />
              <div className="flex-1">
                <Text strong className="text-blue-700">您的问题：</Text>
                <div className="mt-2 text-gray-800">{response.question}</div>
              </div>
            </div>
          </Card>

          {/* AI回答 */}
          <Card 
            title={
              <Space>
                <RobotOutlined className="text-green-500" />
                <span>AI解答</span>
                <Tag color="green">置信度: {Math.round(response.confidence * 100)}%</Tag>
                <Tag color="blue">{response.category}</Tag>
              </Space>
            }
            className="qa-answer-card"
          >
            <div className="space-y-4">
              <Paragraph className="text-base leading-relaxed whitespace-pre-wrap">
                {response.answer}
              </Paragraph>

              {/* 关键词 */}
              {response.keywords.length > 0 && (
                <div>
                  <Text strong className="text-gray-700">关键词：</Text>
                  <div className="mt-2">
                    {response.keywords.map((keyword, index) => (
                      <Tag key={index} color="blue" className="mb-1">
                        {keyword}
                      </Tag>
                    ))}
                  </div>
                </div>
              )}

              {/* 引用来源 */}
              {response.sources.length > 0 && (
                <div>
                  <Text strong className="text-gray-700">引用来源：</Text>
                  <ul className="mt-2 ml-4">
                    {response.sources.map((source, index) => (
                      <li key={index} className="text-gray-600 mb-1">
                        {source}
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          </Card>

          {/* 相关智慧 */}
          {response.related_wisdoms.length > 0 && (
            <Card 
              title={
                <Space>
                  <BookOutlined className="text-cultural-gold" />
                  <span>相关智慧</span>
                </Space>
              }
            >
              <List
                dataSource={response.related_wisdoms}
                renderItem={(wisdom) => (
                  <List.Item className="hover:bg-gray-50 p-4 rounded-lg cursor-pointer">
                    <List.Item.Meta
                      avatar={
                        <div className="w-12 h-12 bg-gradient-to-r from-cultural-gold to-cultural-red rounded-xl flex items-center justify-center">
                          <BookOutlined className="text-white" />
                        </div>
                      }
                      title={
                        <div className="flex items-center justify-between">
                          <span className="font-medium text-gray-800">{wisdom.title}</span>
                          <div className="flex items-center space-x-2">
                            <StarOutlined className="text-yellow-500" />
                            <span className="text-sm text-gray-500">
                              相关度: {Math.round(wisdom.relevance * 100)}%
                            </span>
                          </div>
                        </div>
                      }
                      description={
                        <div className="space-y-2">
                          <div className="flex items-center space-x-4 text-sm text-gray-600">
                            <span>作者：{wisdom.author}</span>
                            <span>学派：{wisdom.school}</span>
                          </div>
                          <Paragraph 
                            className="text-gray-700 mb-0" 
                            ellipsis={{ rows: 2, expandable: true }}
                          >
                            {wisdom.excerpt}
                          </Paragraph>
                        </div>
                      }
                    />
                  </List.Item>
                )}
              />
            </Card>
          )}

          {/* 继续提问 */}
          <Card className="text-center bg-gradient-to-r from-purple-50 to-blue-50">
            <div className="space-y-4">
              <Title level={4} className="text-gray-700">
                还有其他问题吗？
              </Title>
              <Button
                type="primary"
                size="large"
                onClick={() => {
                  setQuestion('');
                  setResponse(null);
                  setError(null);
                }}
                className="px-8"
              >
                继续提问
              </Button>
            </div>
          </Card>
        </div>
      )}
    </div>
  );
};

export default IntelligentQA;