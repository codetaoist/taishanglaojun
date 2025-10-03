import React, { useState } from 'react';
import { Card, Typography, Button, Space, Tabs } from 'antd';
import { MessageOutlined, RobotOutlined, QuestionCircleOutlined, HistoryOutlined } from '@ant-design/icons';
import IntelligentQA from '../components/ai/IntelligentQA';

const { Title, Paragraph } = Typography;

interface ChatMessage {
  id: string;
  content: string;
  sender: 'user' | 'ai';
  timestamp: string;
  type?: 'text' | 'wisdom';
}

interface WisdomRecommendation {
  id: string;
  title: string;
  content: string;
  category: string;
  source: string;
}

const Chat: React.FC = () => {
  const {
    currentConversation,
    messages,
    loading,
    conversationList,
    sessionId,
    sendMessage,
    createNewConversation,
    loadConversation,
    deleteConversation,
    archiveConversation,
    searchConversations,
    exportConversations,
    importConversations,
    retryMessage,
  } = useChat();

  const [inputValue, setInputValue] = useState('');
  const [recommendations, setRecommendations] = useState<WisdomRecommendation[]>([]);
  const [showRecommendations, setShowRecommendations] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const handleSendMessage = async () => {
    if (!inputValue.trim() || loading) return;

    const messageContent = inputValue.trim();
    setInputValue('');

    try {
      await sendMessage(messageContent);
      await fetchWisdomRecommendations(messageContent);
    } catch (error) {
      console.error('发送消息失败:', error);
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSendMessage();
    }
  };

  const handleSelectConversation = (conversationId: string) => {
    loadConversation(conversationId);
  };

  const handleCreateConversation = () => {
    createNewConversation();
  };

  const handleDeleteConversation = (conversationId: string) => {
    deleteConversation(conversationId);
  };

  const handleArchiveConversation = (conversationId: string) => {
    archiveConversation(conversationId);
  };

  const handleSearchConversations = (query: string) => {
    return searchConversations(query);
  };

  const handleExportConversations = () => {
    return exportConversations();
  };

  const handleImportConversations = (data: string) => {
    return importConversations(data);
  };

  // 获取智慧推荐
  const fetchWisdomRecommendations = async (query: string) => {
    try {
      // 使用搜索API来获取相关的文化智慧内容
      const searchResponse = await apiClient.searchWisdom(query, { limit: 3 });
      if (searchResponse.success && searchResponse.data.items.length > 0) {
        const wisdomRecommendations: WisdomRecommendation[] = searchResponse.data.items.map(item => ({
          wisdom_id: item.id,
          title: item.title,
          author: item.author,
          category: item.category,
          summary: item.summary,
          relevance: 0.8,
          reason: '与您的问题相关'
        }));
        setRecommendations(wisdomRecommendations);
        setShowRecommendations(true);
      }
    } catch (error) {
      console.error('Failed to fetch wisdom recommendations:', error);
    }
  };

  // 处理智慧点击，获取AI解读
  const handleWisdomClick = async (wisdomId: string, title: string) => {
    try {
      const response = await apiClient.getWisdomInterpretation(wisdomId);
      if (response.success) {
        const newMessage: ChatMessage = {
          id: Date.now().toString(),
          content: `关于"${title}"的AI解读：\n\n${response.data.interpretation}`,
          sender: 'ai',
          timestamp: new Date(),
          metadata: {
            wisdomId,
            wisdomTitle: title,
            category: 'interpretation'
          }
        };
        setMessages(prev => [...prev, newMessage]);
      }
    } catch (error) {
      console.error('Failed to get wisdom interpretation:', error);
      message.error('获取智慧解读失败');
    }
  };

  const quickQuestions = [
    '什么是修身养性？',
    '如何理解"知行合一"？',
    '道德经的核心思想是什么？',
    '儒家的仁义礼智信如何理解？'
  ];

  const handleQuickQuestion = (question: string) => {
    setInputValue(question);
  };

  const chatModeItems: MenuProps['items'] = [
    {
      key: 'wisdom',
      label: '智慧问答模式',
      icon: <BookOutlined />,
    },
    {
      key: 'interpretation',
      label: '经典解读模式',
      icon: <BulbOutlined />,
    },
    {
      key: 'general',
      label: '通用对话模式',
      icon: <RobotOutlined />,
    },
  ];

  const clearChat = () => {
    setMessages([]);
    setSessionId(undefined);
    setRecommendations([]);
    setShowRecommendations(false);
    message.success('对话已清空');
  };

  const renderMessage = (message: ChatMessage) => {
    const isUser = message.role === 'user';
    
    return (
      <div key={message.id} className={`flex ${isUser ? 'justify-end' : 'justify-start'} mb-4`}>
        <div className={`flex max-w-[80%] ${isUser ? 'flex-row-reverse' : 'flex-row'} items-start space-x-3`}>
          <Avatar 
            icon={isUser ? <UserOutlined /> : <RobotOutlined />}
            className={isUser ? 'bg-primary-500 ml-3' : 'bg-cultural-gold mr-3'}
          />
          <div className={`rounded-lg p-4 ${
            isUser 
              ? 'bg-primary-500 text-white' 
              : 'bg-white border border-gray-200 shadow-sm'
          }`}>
            {message.metadata?.wisdomTitle && (
              <div className="mb-2 pb-2 border-b border-gray-200">
                <Tag color="gold" icon={<BookOutlined />}>
                  {message.metadata.wisdomTitle}
                </Tag>
              </div>
            )}
            
            <Paragraph 
              className={`mb-0 ${isUser ? 'text-white' : 'text-gray-800'}`}
              style={{ whiteSpace: 'pre-wrap' }}
            >
              {message.content}
            </Paragraph>
            
            {message.metadata?.sources && (
              <div className="mt-2 pt-2 border-t border-gray-200">
                <span className="text-xs text-gray-500">
                  参考来源: {message.metadata.sources.join(', ')}
                </span>
                {message.metadata.confidence && (
                  <span className="ml-2 text-xs text-gray-500">
                    可信度: {(message.metadata.confidence * 100).toFixed(0)}%
                  </span>
                )}
              </div>
            )}
            
            <div className="mt-2">
              <span 
                className={`text-xs ${isUser ? 'text-blue-100' : 'text-gray-500'}`}
              >
                {new Date(message.timestamp).toLocaleTimeString()}
              </span>
            </div>
          </div>
        </div>
      </div>
    );
  };

  return (
    <Layout style={{ height: '100vh' }}>
      <Row style={{ height: '100%' }}>
        {/* 左侧对话列表 */}
        <Col span={6} style={{ borderRight: '1px solid #f0f0f0', height: '100%', overflow: 'hidden' }}>
          <ConversationList
            conversations={conversationList}
            currentConversationId={currentConversation?.id}
            loading={loading}
            onSelectConversation={handleSelectConversation}
            onCreateConversation={handleCreateConversation}
            onDeleteConversation={handleDeleteConversation}
            onArchiveConversation={handleArchiveConversation}
            onSearchConversations={handleSearchConversations}
            onExportConversations={handleExportConversations}
            onImportConversations={handleImportConversations}
          />
        </Col>

        {/* 右侧聊天区域 */}
        <Col span={18}>
          <Content style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
            {/* 聊天标题 */}
            <Card 
              size="small" 
              style={{ 
                borderRadius: 0, 
                borderBottom: '1px solid #f0f0f0',
                flexShrink: 0
              }}
            >
              <Space align="center">
                <RobotOutlined style={{ fontSize: '20px', color: '#1890ff' }} />
                <Title level={4} style={{ margin: 0 }}>
                  {currentConversation?.title || '太上老君智慧助手'}
                </Title>
                {sessionId && (
                  <Tag color="blue" size="small">
                    会话: {sessionId.slice(-8)}
                  </Tag>
                )}
              </Space>
            </Card>

            {/* 消息列表 */}
            <div style={{ 
              flex: 1, 
              overflow: 'auto', 
              padding: '16px',
              backgroundColor: '#fafafa'
            }}>
              {messages.length === 0 ? (
                <div style={{ 
                  textAlign: 'center', 
                  marginTop: '100px',
                  color: '#999'
                }}>
                  <RobotOutlined style={{ fontSize: '48px', marginBottom: '16px' }} />
                  <Title level={3} type="secondary">
                    欢迎使用太上老君智慧助手
                  </Title>
                  <Text type="secondary">
                    我可以为您提供传统文化智慧指导，请开始对话吧！
                  </Text>
                </div>
              ) : (
                <List
                  dataSource={messages}
                  renderItem={(message) => (
                    <List.Item
                      key={message.id}
                      style={{
                        border: 'none',
                        padding: '8px 0',
                        justifyContent: message.sender === 'user' ? 'flex-end' : 'flex-start',
                      }}
                    >
                      <div
                        style={{
                          maxWidth: '70%',
                          display: 'flex',
                          flexDirection: message.sender === 'user' ? 'row-reverse' : 'row',
                          alignItems: 'flex-start',
                          gap: '8px',
                        }}
                      >
                        <Avatar
                          icon={message.sender === 'user' ? <UserOutlined /> : <RobotOutlined />}
                          style={{
                            backgroundColor: message.sender === 'user' ? '#1890ff' : '#52c41a',
                            flexShrink: 0,
                          }}
                        />
                        <Card
                          size="small"
                          style={{
                            backgroundColor: message.sender === 'user' ? '#1890ff' : '#fff',
                            color: message.sender === 'user' ? '#fff' : '#000',
                            borderRadius: '12px',
                            border: message.sender === 'user' ? 'none' : '1px solid #f0f0f0',
                            boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
                          }}
                          bodyStyle={{ padding: '8px 12px' }}
                        >
                          <Text
                            style={{
                              color: message.sender === 'user' ? '#fff' : '#000',
                              whiteSpace: 'pre-wrap',
                            }}
                          >
                            {message.content}
                          </Text>
                          <div style={{ 
                            marginTop: '4px', 
                            fontSize: '11px', 
                            opacity: 0.7,
                            textAlign: message.sender === 'user' ? 'right' : 'left'
                          }}>
                            {new Date(message.timestamp).toLocaleTimeString()}
                          </div>
                        </Card>
                      </div>
                    </List.Item>
                  )}
                />
              )}
              
              {loading && (
                <div style={{ textAlign: 'center', padding: '16px' }}>
                  <Spin size="small" />
                  <Text style={{ marginLeft: '8px', color: '#999' }}>
                    AI正在思考中...
                  </Text>
                </div>
              )}
              
              <div ref={messagesEndRef} />
            </div>

            {/* 智慧推荐区域 */}
            {showRecommendations && recommendations.length > 0 && (
              <Card
                size="small"
                title={
                  <Space>
                    <BulbOutlined style={{ color: '#faad14' }} />
                    <Text>相关文化智慧推荐</Text>
                  </Space>
                }
                extra={
                  <Button
                    type="text"
                    size="small"
                    onClick={() => setShowRecommendations(false)}
                  >
                    收起
                  </Button>
                }
                style={{ 
                  margin: '0 16px',
                  borderRadius: '8px',
                  flexShrink: 0
                }}
              >
                <Space direction="vertical" style={{ width: '100%' }}>
                  {recommendations.map((rec) => (
                    <Card
                      key={rec.id}
                      size="small"
                      hoverable
                      style={{ backgroundColor: '#f9f9f9' }}
                    >
                      <Space direction="vertical" size="small" style={{ width: '100%' }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                          <Text strong>{rec.title}</Text>
                          <Tag color="blue" size="small">{rec.category}</Tag>
                        </div>
                        <Text type="secondary" style={{ fontSize: '12px' }}>
                          {rec.content}
                        </Text>
                        <Text type="secondary" style={{ fontSize: '11px' }}>
                          来源：{rec.source}
                        </Text>
                      </Space>
                    </Card>
                  ))}
                </Space>
              </Card>
            )}

            {/* 输入区域 */}
            <Card 
              size="small" 
              style={{ 
                borderRadius: 0, 
                borderTop: '1px solid #f0f0f0',
                flexShrink: 0
              }}
            >
              <Space.Compact style={{ width: '100%' }}>
                <TextArea
                  value={inputValue}
                  onChange={(e) => setInputValue(e.target.value)}
                  onKeyPress={handleKeyPress}
                  placeholder="请输入您的问题..."
                  autoSize={{ minRows: 1, maxRows: 4 }}
                  style={{ resize: 'none' }}
                  disabled={loading}
                />
                <Button
                  type="primary"
                  icon={<SendOutlined />}
                  onClick={handleSendMessage}
                  loading={loading}
                  disabled={!inputValue.trim()}
                  style={{ height: 'auto' }}
                >
                  发送
                </Button>
              </Space.Compact>
            </Card>
          </Content>
        </Col>
      </Row>
    </Layout>
  );
};

export default Chat;