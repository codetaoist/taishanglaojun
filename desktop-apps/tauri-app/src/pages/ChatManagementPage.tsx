import React, { useState, useEffect, useRef } from 'react';
import {
  MessageCircle,
  Send,
  Plus,
  MoreVertical,
  Search,
  RefreshCw,
  Users,
  User,
  Paperclip,
  Smile,
  Info,
  Image,
  Video,
  Phone,
  Wifi,
  WifiOff,
  TestTube,
} from 'lucide-react';
import chatService, { Chat, ChatMessage, CreateChatRequest } from '../services/chatService';
import { chatTester, TestResult } from '../utils/chatTest';

const ChatManagementPage: React.FC = () => {
  const [chats, setChats] = useState<Chat[]>([]);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [selectedChat, setSelectedChat] = useState<Chat | null>(null);
  const [newMessage, setNewMessage] = useState('');
  const [searchTerm, setSearchTerm] = useState('');
  const [showChatDialog, setShowChatDialog] = useState(false);
  const [isConnected, setIsConnected] = useState(false);
  const [connectionStatus, setConnectionStatus] = useState<'connecting' | 'connected' | 'disconnected'>('disconnected');
  const [isLoading, setIsLoading] = useState(false);
  const [newChatName, setNewChatName] = useState('');
  const [newChatType, setNewChatType] = useState<'Private' | 'Group'>('Private');
  const [isTestingChat, setIsTestingChat] = useState(false);
  const [testResults, setTestResults] = useState<TestResult[]>([]);
  const [showTestResults, setShowTestResults] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // 初始化聊天服务
  useEffect(() => {
    const initializeChatService = async () => {
      try {
        setIsLoading(true);
        
        // 设置当前用户（临时使用固定用户ID）
        await chatService.setCurrentUser('current_user_123');
        
        // 检查连接状态
        const connected = await chatService.checkConnectionStatus();
        setIsConnected(connected);
        
        // 如果未连接，尝试连接WebSocket
        if (!connected) {
          try {
            await chatService.connectWebSocket();
            setIsConnected(true);
          } catch (error) {
            console.warn('WebSocket连接失败，将使用缓存数据:', error);
          }
        }
        
        // 加载聊天列表
        await loadChatList();
        
      } catch (error) {
        console.error('初始化聊天服务失败:', error);
      } finally {
        setIsLoading(false);
      }
    };

    initializeChatService();

    // 添加连接状态监听器
    const connectionListener = (connected: boolean) => {
      setIsConnected(connected);
    };
    chatService.addConnectionListener(connectionListener);

    // 添加消息监听器
    const messageListener = (message: ChatMessage) => {
      setMessages(prev => [...prev, message]);
    };
    chatService.addMessageListener(messageListener);

    // 清理函数
    return () => {
      chatService.removeConnectionListener(connectionListener);
      chatService.removeMessageListener(messageListener);
    };
  }, []);

  // 加载聊天列表
  const loadChatList = async () => {
    try {
      let chatList: Chat[] = [];
      
      if (isConnected) {
        // 如果连接正常，从服务器获取最新数据
        const response = await chatService.getChatList();
        if (response.success && response.chats) {
          chatList = response.chats;
        }
      } else {
        // 如果未连接，使用缓存数据
        chatList = await chatService.getCachedChatList();
      }
      
      // 如果没有数据，使用模拟数据
      if (chatList.length === 0) {
        chatList = [
          {
            id: '1',
            name: '开发团队',
            type: 'group',
            lastMessage: '新版本已经发布了',
            lastMessageTime: '09:45',
            unreadCount: 0,
            isOnline: false,
            participants: ['张三', '李四', '王五'],
          },
          {
            id: '2',
            name: '客户支持',
            type: 'private',
            lastMessage: '感谢您的反馈',
            lastMessageTime: '昨天',
            unreadCount: 1,
            isOnline: true,
          },
        ];
      }
      
      setChats(chatList);
      if (chatList.length > 0 && !selectedChat) {
        setSelectedChat(chatList[0]);
      }
    } catch (error) {
      console.error('加载聊天列表失败:', error);
    }
  };

  // 加载聊天消息
  const loadChatMessages = async (chatId: string) => {
    try {
      let messageList: ChatMessage[] = [];
      
      if (isConnected) {
        // 如果连接正常，从服务器获取最新数据
        const response = await chatService.getChatMessages(chatId);
        if (response.success && response.messages) {
          messageList = response.messages;
        }
      } else {
        // 如果未连接，使用缓存数据
        messageList = await chatService.getCachedMessages(chatId);
      }
      
      // 如果没有数据，使用模拟数据
      if (messageList.length === 0) {
        messageList = [
          {
            id: '1',
            chatId: chatId,
            senderId: 'other_user',
            senderName: '团队成员',
            content: '大家好！欢迎使用聊天功能',
            timestamp: '10:25',
            type: 'text',
            isRead: true,
          },
          {
            id: '2',
            chatId: chatId,
            senderId: 'current_user_123',
            senderName: '我',
            content: '谢谢！这个功能很棒',
            timestamp: '10:28',
            type: 'text',
            isRead: true,
          },
        ];
      }
      
      setMessages(messageList);
    } catch (error) {
      console.error('加载聊天消息失败:', error);
    }
  };

  // 当选择聊天时加载消息
  useEffect(() => {
    if (selectedChat) {
      loadChatMessages(selectedChat.id);
    }
  }, [selectedChat, isConnected]);

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const handleSendMessage = async () => {
    if (!newMessage.trim() || !selectedChat) return;

    const tempMessage: ChatMessage = {
      id: Date.now().toString(),
      chatId: selectedChat.id,
      senderId: 'current_user_123',
      senderName: '我',
      content: newMessage,
      timestamp: new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }),
      type: 'text',
      isRead: true,
    };

    // 立即显示消息
    setMessages(prev => [...prev, tempMessage]);
    const messageContent = newMessage;
    setNewMessage('');

    try {
      // 发送消息到服务器
      await chatService.sendMessage({
        chat_id: selectedChat.id,
        content: messageContent,
        message_type: 'Text',
      });
      
      console.log('消息发送成功');
    } catch (error) {
      console.error('发送消息失败:', error);
      // 可以在这里添加错误提示或重试机制
    }
  };

  const handleCreateChat = async () => {
    if (!newChatName.trim()) return;

    setIsLoading(true);
    try {
      const request: CreateChatRequest = {
        chat_type: newChatType,
        name: newChatName,
        participants: []
      };

      const response = await chatService.createChat(request);
      if (response.success) {
        setShowChatDialog(false);
        setNewChatName('');
        setNewChatType('Private');
        await loadChatList(); // 重新加载聊天列表
      } else {
        console.error('创建聊天失败:', response.message);
      }
    } catch (error) {
      console.error('创建聊天时出错:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleRunChatTests = async () => {
    setIsTestingChat(true);
    setShowTestResults(true);
    try {
      const results = await chatTester.runAllTests();
      setTestResults(results);
    } catch (error) {
      console.error('运行测试时出错:', error);
    } finally {
      setIsTestingChat(false);
    }
  };

  const handleRefresh = async () => {
    setIsLoading(true);
    try {
      await loadChatList();
      if (selectedChat) {
        await loadChatMessages(selectedChat.id);
      }
    } finally {
      setIsLoading(false);
    }
  };

  const handleConnectWebSocket = async () => {
    try {
      await chatService.connectWebSocket();
      setIsConnected(true);
    } catch (error) {
      console.error('连接WebSocket失败:', error);
    }
  };

  const filteredChats = chats.filter(chat =>
    chat.name.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const currentMessages = messages.filter(msg => msg.chatId === selectedChat?.id);

  return (
    <div className="flex h-screen bg-background">
      {/* 左侧聊天列表 */}
      <div className="w-80 border-r border-border bg-card">
        {/* 头部 */}
        <div className="p-4 border-b border-border">
          <div className="flex items-center justify-between mb-4">
            <h1 className="text-xl font-bold text-foreground">聊天</h1>
            <div className="flex items-center space-x-2">
              {/* 连接状态指示器 */}
              <div className="flex items-center space-x-1">
                {isConnected ? (
                  <div title="已连接">
                    <Wifi className="h-4 w-4 text-green-500" />
                  </div>
                ) : (
                  <div 
                    title="未连接 - 点击重连"
                    onClick={handleConnectWebSocket}
                    className="cursor-pointer"
                  >
                    <WifiOff className="h-4 w-4 text-red-500" />
                  </div>
                )}
              </div>
              
              <button
                onClick={handleRefresh}
                disabled={isLoading}
                className="p-2 text-muted-foreground hover:text-foreground hover:bg-secondary rounded-lg transition-colors disabled:opacity-50"
                title="刷新"
              >
                <RefreshCw className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
              </button>
              <button
                onClick={handleRunChatTests}
                disabled={isTestingChat}
                className="p-2 text-muted-foreground hover:text-foreground hover:bg-secondary rounded-lg transition-colors disabled:opacity-50"
                title="运行聊天功能测试"
              >
                {isTestingChat ? (
                  <RefreshCw className="h-4 w-4 animate-spin text-blue-500" />
                ) : (
                  <TestTube className="h-4 w-4" />
                )}
              </button>
              <button
                onClick={() => setShowChatDialog(true)}
                className="p-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors"
                title="新建聊天"
              >
                <Plus className="h-4 w-4" />
              </button>
            </div>
          </div>
          
          {/* 搜索框 */}
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <input
              type="text"
              placeholder="搜索聊天..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full pl-10 pr-4 py-2 border border-border rounded-lg bg-background text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
            />
          </div>
        </div>

        {/* 聊天列表 */}
        <div className="overflow-y-auto">
          {filteredChats.map((chat) => (
            <div
              key={chat.id}
              onClick={() => setSelectedChat(chat)}
              className={`p-4 border-b border-border cursor-pointer hover:bg-secondary/50 transition-colors ${
                selectedChat?.id === chat.id ? 'bg-secondary' : ''
              }`}
            >
              <div className="flex items-center space-x-3">
                <div className="relative">
                  <div className="w-12 h-12 bg-primary/10 rounded-full flex items-center justify-center">
                    {chat.type === 'group' ? (
                      <Users className="h-6 w-6 text-primary" />
                    ) : (
                      <User className="h-6 w-6 text-primary" />
                    )}
                  </div>
                  {chat.isOnline && (
                    <div className="absolute bottom-0 right-0 w-3 h-3 bg-green-500 rounded-full border-2 border-background"></div>
                  )}
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between">
                    <h3 className="text-sm font-medium text-foreground truncate">{chat.name}</h3>
                    <span className="text-xs text-muted-foreground">{chat.lastMessageTime}</span>
                  </div>
                  <p className="text-sm text-muted-foreground truncate mt-1">{chat.lastMessage}</p>
                  {chat.type === 'group' && chat.participants && (
                    <p className="text-xs text-muted-foreground mt-1">
                      {chat.participants.length} 位成员
                    </p>
                  )}
                </div>
                {chat.unreadCount > 0 && (
                  <div className="w-5 h-5 bg-primary text-primary-foreground rounded-full flex items-center justify-center text-xs font-medium">
                    {chat.unreadCount}
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* 右侧聊天区域 */}
      <div className="flex-1 flex flex-col">
        {selectedChat ? (
          <>
            {/* 聊天头部 */}
            <div className="p-4 border-b border-border bg-card">
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-3">
                  <div className="w-10 h-10 bg-primary/10 rounded-full flex items-center justify-center">
                    {selectedChat.type === 'group' ? (
                      <Users className="h-5 w-5 text-primary" />
                    ) : (
                      <User className="h-5 w-5 text-primary" />
                    )}
                  </div>
                  <div>
                    <h2 className="text-lg font-semibold text-foreground">{selectedChat.name}</h2>
                    <p className="text-sm text-muted-foreground">
                      {selectedChat.isOnline ? '在线' : '离线'}
                      {selectedChat.type === 'group' && selectedChat.participants && 
                        ` • ${selectedChat.participants.length} 位成员`
                      }
                    </p>
                  </div>
                </div>
                <div className="flex items-center space-x-2">
                  <button className="p-2 text-muted-foreground hover:text-foreground hover:bg-secondary rounded-lg transition-colors">
                    <Phone className="h-4 w-4" />
                  </button>
                  <button className="p-2 text-muted-foreground hover:text-foreground hover:bg-secondary rounded-lg transition-colors">
                    <Video className="h-4 w-4" />
                  </button>
                  <button className="p-2 text-muted-foreground hover:text-foreground hover:bg-secondary rounded-lg transition-colors">
                    <Info className="h-4 w-4" />
                  </button>
                  <button className="p-2 text-muted-foreground hover:text-foreground hover:bg-secondary rounded-lg transition-colors">
                    <MoreVertical className="h-4 w-4" />
                  </button>
                </div>
              </div>
            </div>

            {/* 消息区域 */}
            <div className="flex-1 overflow-y-auto p-4 space-y-4">
              {currentMessages.map((message) => (
                <div
                  key={message.id}
                  className={`flex ${message.senderId === 'current_user_123' ? 'justify-end' : 'justify-start'}`}
                >
                  <div
                    className={`max-w-xs lg:max-w-md px-4 py-2 rounded-lg ${
                      message.senderId === 'current_user_123'
                        ? 'bg-primary text-primary-foreground'
                        : 'bg-secondary text-secondary-foreground'
                    }`}
                  >
                    {message.senderId !== 'current_user_123' && (
                      <p className="text-xs font-medium mb-1">{message.senderName}</p>
                    )}
                    <p className="text-sm">{message.content}</p>
                    <p className={`text-xs mt-1 ${
                      message.senderId === 'current_user_123' 
                        ? 'text-primary-foreground/70' 
                        : 'text-muted-foreground'
                    }`}>
                      {message.timestamp}
                    </p>
                  </div>
                </div>
              ))}
              <div ref={messagesEndRef} />
            </div>

            {/* 输入区域 */}
            <div className="p-4 border-t border-border bg-card">
              <div className="flex items-center space-x-2">
                <button className="p-2 text-muted-foreground hover:text-foreground hover:bg-secondary rounded-lg transition-colors">
                  <Paperclip className="h-4 w-4" />
                </button>
                <button className="p-2 text-muted-foreground hover:text-foreground hover:bg-secondary rounded-lg transition-colors">
                  <Image className="h-4 w-4" />
                </button>
                <button className="p-2 text-muted-foreground hover:text-foreground hover:bg-secondary rounded-lg transition-colors">
                  <Smile className="h-4 w-4" />
                </button>
                <div className="flex-1 relative">
                  <input
                    type="text"
                    placeholder="输入消息..."
                    value={newMessage}
                    onChange={(e) => setNewMessage(e.target.value)}
                    onKeyPress={(e) => e.key === 'Enter' && handleSendMessage()}
                    className="w-full px-4 py-2 border border-border rounded-lg bg-background text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
                  />
                </div>
                <button
                  onClick={handleSendMessage}
                  disabled={!newMessage.trim()}
                  className="p-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <Send className="h-4 w-4" />
                </button>
              </div>
            </div>
          </>
        ) : (
          <div className="flex-1 flex items-center justify-center">
            <div className="text-center">
              <MessageCircle className="h-16 w-16 text-muted-foreground mx-auto mb-4" />
              <h3 className="text-lg font-medium text-foreground mb-2">选择一个聊天</h3>
              <p className="text-muted-foreground">从左侧列表中选择一个聊天开始对话</p>
            </div>
          </div>
        )}
      </div>

      {/* 新建聊天对话框 */}
      {showChatDialog && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-card border border-border rounded-lg p-6 w-full max-w-md">
            <h2 className="text-lg font-semibold text-foreground mb-4">新建聊天</h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-foreground mb-2">聊天名称</label>
                <input
                  type="text"
                  value={newChatName}
                  onChange={(e) => setNewChatName(e.target.value)}
                  className="w-full px-3 py-2 border border-border rounded-lg bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
                  placeholder="输入聊天名称"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-foreground mb-2">聊天类型</label>
                <select 
                  value={newChatType}
                  onChange={(e) => setNewChatType(e.target.value as 'Private' | 'Group')}
                  className="w-full px-3 py-2 border border-border rounded-lg bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
                >
                  <option value="Private">私人聊天</option>
                  <option value="Group">群组聊天</option>
                </select>
              </div>
            </div>
            <div className="flex justify-end space-x-3 mt-6">
              <button
                onClick={() => {
                  setShowChatDialog(false);
                  setNewChatName('');
                  setNewChatType('Private');
                }}
                className="px-4 py-2 text-muted-foreground hover:text-foreground transition-colors"
              >
                取消
              </button>
              <button
                onClick={handleCreateChat}
                disabled={!newChatName.trim()}
                className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              >
                创建
              </button>
            </div>
          </div>
        </div>
      )}

      {/* 测试结果对话框 */}
      {showTestResults && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-card border border-border rounded-lg p-6 w-[600px] max-w-[90vw] max-h-[80vh] overflow-hidden flex flex-col">
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-lg font-semibold text-foreground">聊天功能测试结果</h3>
              <button
                onClick={() => setShowTestResults(false)}
                className="text-muted-foreground hover:text-foreground"
              >
                ✕
              </button>
            </div>
            
            <div className="flex-1 overflow-y-auto">
              {isTestingChat ? (
                <div className="flex items-center justify-center py-8">
                  <RefreshCw className="w-8 h-8 animate-spin text-primary mr-3" />
                  <span className="text-muted-foreground">正在运行测试...</span>
                </div>
              ) : testResults.length > 0 ? (
                <div className="space-y-3">
                  {/* 测试摘要 */}
                  <div className="bg-secondary p-4 rounded-lg">
                    <div className="grid grid-cols-3 gap-4 text-center">
                      <div>
                        <div className="text-2xl font-bold text-foreground">{testResults.length}</div>
                        <div className="text-sm text-muted-foreground">总测试数</div>
                      </div>
                      <div>
                        <div className="text-2xl font-bold text-green-600">
                          {testResults.filter(r => r.success).length}
                        </div>
                        <div className="text-sm text-muted-foreground">通过</div>
                      </div>
                      <div>
                        <div className="text-2xl font-bold text-red-600">
                          {testResults.filter(r => !r.success).length}
                        </div>
                        <div className="text-sm text-muted-foreground">失败</div>
                      </div>
                    </div>
                  </div>
                  
                  {/* 详细测试结果 */}
                  <div className="space-y-2">
                    {testResults.map((result, index) => (
                      <div
                        key={index}
                        className={`p-3 rounded-lg border ${
                          result.success
                            ? 'bg-green-50 border-green-200 dark:bg-green-950 dark:border-green-800'
                            : 'bg-red-50 border-red-200 dark:bg-red-950 dark:border-red-800'
                        }`}
                      >
                        <div className="flex items-center justify-between">
                          <div className="flex items-center">
                            <span className={`mr-2 ${result.success ? 'text-green-600' : 'text-red-600'}`}>
                              {result.success ? '✅' : '❌'}
                            </span>
                            <span className="font-medium text-foreground">{result.testName}</span>
                          </div>
                          <span className="text-sm text-muted-foreground">{result.duration}ms</span>
                        </div>
                        {!result.success && (
                          <div className="mt-2 text-sm text-red-600 ml-6">
                            {result.message}
                          </div>
                        )}
                      </div>
                    ))}
                  </div>
                </div>
              ) : (
                <div className="text-center py-8 text-muted-foreground">
                  暂无测试结果
                </div>
              )}
            </div>
            
            <div className="flex justify-end space-x-3 mt-4 pt-4 border-t border-border">
              <button
                onClick={() => setShowTestResults(false)}
                className="px-4 py-2 text-muted-foreground hover:text-foreground transition-colors"
              >
                关闭
              </button>
              <button
                onClick={handleRunChatTests}
                disabled={isTestingChat}
                className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                {isTestingChat ? '测试中...' : '重新测试'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default ChatManagementPage;