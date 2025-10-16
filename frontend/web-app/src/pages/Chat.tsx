import React, { useState, useRef, useEffect } from 'react';
import { 
  Layout, 
  Card, 
  Typography, 
  Button, 
  Space, 
  Input, 
  List, 
  Avatar, 
  Tag, 
  Spin, 
  Row, 
  Col 
} from 'antd';
import { getNotificationInstance } from '../services/notificationService';
import { 
  MessageOutlined, 
  RobotOutlined, 
  UserOutlined, 
  SendOutlined, 
  BulbOutlined, 
  BookOutlined 
} from '@ant-design/icons';
import { useChat } from '../hooks/useChat';
import ConversationList from '../components/chat/ConversationList';
import { apiClient } from '../services/api';
import './Chat.css';

const { Title, Text, Paragraph } = Typography;
const { Content } = Layout;
const { TextArea } = Input;

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
    conversations,
    sessionId,
    sendMessage,
    createNewConversation,
    loadConversation,
    deleteConversation,
    archiveConversation,
    searchConversations,
    exportConversations,
    importConversations,
  } = useChat();

  const [inputValue, setInputValue] = useState('');
  const [recommendations, setRecommendations] = useState<WisdomRecommendation[]>([]);
  const [showRecommendations, setShowRecommendations] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // 自动滚动到底部
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const handleSendMessage = async () => {
    if (!inputValue.trim() || loading) return;
    
    const messageContent = inputValue.trim();
    setInputValue('');
    
    try {
      await sendMessage(messageContent);
      // 获取智慧推荐
      await fetchWisdomRecommendations(messageContent);
    } catch (error) {
      console.error('Failed to send message:', error);
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
      // 模拟API调用获取智慧推荐
      const mockRecommendations: WisdomRecommendation[] = [
        {
          id: '1',
          title: '道德经第一章',
          content: '道可道，非常道。名可名，非常名。',
          category: '道家经典',
          source: '老子'
        },
        {
          id: '2',
          title: '论语·学而',
          content: '学而时习之，不亦说乎？',
          category: '儒家经典',
          source: '孔子'
        }
      ];
      
      setRecommendations(mockRecommendations);
      setShowRecommendations(true);
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
      const notification = getNotificationInstance();
      notification.error({
        message: '错误',
        description: '获取智慧解读失败'
      });
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
    const notification = getNotificationInstance();
    notification.success({
      message: '成功',
      description: '对话已清空'
    });
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
    <div className="chat-container">
      <Layout className="chat-layout">
        <Row style={{ height: '100%' }}>
        {/* 左侧对话列表 */}
        <Col span={6} className="conversation-sidebar">
          <ConversationList
            conversations={conversations}
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
        <Col span={18} className="chat-main">
          <Content style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
            {/* 聊天标题 */}
            <Card 
              size="small" 
              className="chat-header-card"
            >
              <Space align="center">
                <RobotOutlined className="chat-header-icon" style={{ fontSize: '20px' }} />
                <Title level={4} className="chat-header-title">
                  {currentConversation?.title || '太上老君智慧助手'}
                </Title>
                {sessionId && (
                  <Tag className="chat-tag" size="small">
                    会话: {sessionId.slice(-8)}
                  </Tag>
                )}
              </Space>
            </Card>

            {/* 消息列表 */}
            <div className="chat-messages-area">
              {messages.length === 0 ? (
                <div className="chat-empty-state">
                  <RobotOutlined className="chat-empty-icon" style={{ fontSize: '48px', marginBottom: '16px' }} />
                  <Title level={3} className="chat-empty-title">
                    欢迎使用太上老君智慧助手
                  </Title>
                  <Text className="chat-empty-text">
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
                        justifyContent: message.role === 'user' ? 'flex-end' : 'flex-start',
                      }}
                    >
                      <div
                        style={{
                          maxWidth: '70%',
                          display: 'flex',
                          flexDirection: message.role === 'user' ? 'row-reverse' : 'row',
                          alignItems: 'flex-start',
                          gap: '8px',
                        }}
                      >
                        <Avatar
                          icon={message.role === 'user' ? <UserOutlined /> : <RobotOutlined />}
                          className="message-avatar"
                          style={{
                            backgroundColor: message.role === 'user' ? '#1890ff' : '#52c41a',
                            flexShrink: 0,
                          }}
                        />
                        <Card
                          size="small"
                          className={`message-card ${message.role === 'user' ? 'user-message-card' : 'ai-message-card'}`}
                          styles={{ body: { padding: '8px 12px' } }}
                        >
                          <Text
                            className={message.role === 'user' ? 'user-message-text' : 'ai-message-text'}
                            style={{
                              whiteSpace: 'pre-wrap',
                            }}
                          >
                            {message.content}
                          </Text>
                          <div 
                            className="message-timestamp"
                            style={{ 
                              marginTop: '4px', 
                              fontSize: '11px', 
                              textAlign: message.role === 'user' ? 'right' : 'left'
                            }}
                          >
                            {new Date(message.timestamp).toLocaleTimeString()}
                          </div>
                        </Card>
                      </div>
                    </List.Item>
                  )}
                />
              )}
              
              {loading && (
                <div className="chat-loading">
                  <Spin size="small" />
                  <Text className="chat-loading-text" style={{ marginLeft: '8px' }}>
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
                className="wisdom-recommendations-card"
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
                    className="chat-button"
                    onClick={() => setShowRecommendations(false)}
                  >
                    收起
                  </Button>
                }
              >
                <Space direction="vertical" style={{ width: '100%' }}>
                  {recommendations.map((rec) => (
                    <Card
                      key={rec.id}
                      size="small"
                      hoverable
                      className="wisdom-recommendation-item"
                      onClick={() => {
                        setInputValue(rec.content);
                        setShowRecommendations(false);
                      }}
                    >
                      <Space direction="vertical" size="small" style={{ width: '100%' }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                          <Text strong className="wisdom-recommendation-title">{rec.title}</Text>
                          <Tag className="chat-tag" size="small">{rec.category}</Tag>
                        </div>
                        <Text className="wisdom-recommendation-content" style={{ fontSize: '12px' }}>
                          {rec.content}
                        </Text>
                        <Text className="wisdom-recommendation-source" style={{ fontSize: '11px' }}>
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
              className="chat-input-card"
            >
              <Space.Compact style={{ width: '100%' }}>
                <TextArea
                  value={inputValue}
                  onChange={(e) => setInputValue(e.target.value)}
                  onKeyPress={handleKeyPress}
                  placeholder="请输入您的问题..."
                  autoSize={{ minRows: 1, maxRows: 4 }}
                  className="chat-input-area"
                  style={{ resize: 'none' }}
                  disabled={loading}
                />
                <Button
                  type="primary"
                  icon={<SendOutlined />}
                  onClick={handleSendMessage}
                  loading={loading}
                  disabled={!inputValue.trim()}
                  className="chat-send-button"
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
    </div>
  );
};

export default Chat;