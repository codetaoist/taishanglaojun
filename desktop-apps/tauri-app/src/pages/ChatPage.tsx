import { useState, useRef, useEffect } from 'react';
import { Bot, User, Loader2, Plus, Trash2, Settings, X } from 'lucide-react';
import { cn } from '../utils/cn';
import aiService, { ChatMessage, ChatSession } from '../services/aiService';
import InputOptimizer from '../components/InputOptimizer';

export default function ChatPage() {
  const [sessions, setSessions] = useState<ChatSession[]>([]);
  const [currentSession, setCurrentSession] = useState<ChatSession | null>(null);
  const [message, setMessage] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [selectedModel, setSelectedModel] = useState('taishanglaojun-chat');
  const [selectedProvider, setSelectedProvider] = useState('local');
  const [showSettings, setShowSettings] = useState(false);
  const [showSidebar, setShowSidebar] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    loadChatSessions();
  }, []);

  useEffect(() => {
    scrollToBottom();
  }, [currentSession?.messages]);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const loadChatSessions = async () => {
    try {
      // 使用新的AI服务获取会话
      const mockSessions: ChatSession[] = [
        {
          id: 'session_1',
          title: '新对话',
          messages: [],
          createdAt: new Date(),
          updatedAt: new Date(),
          model: selectedModel,
          provider: selectedProvider
        }
      ];
      setSessions(mockSessions);
      if (mockSessions.length > 0) {
        setCurrentSession(mockSessions[0]);
      }
    } catch (error) {
      console.error('Failed to load chat sessions:', error);
    }
  };

  const createNewSession = async () => {
    try {
      const newSession: ChatSession = {
        id: `session_${Date.now()}`,
        title: '新对话',
        messages: [],
        createdAt: new Date(),
        updatedAt: new Date(),
        model: selectedModel,
        provider: selectedProvider
      };

      setSessions(prev => [newSession, ...prev]);
      setCurrentSession(newSession);
    } catch (error) {
      console.error('Failed to create new session:', error);
    }
  };

  const deleteSession = async (sessionId: string) => {
    try {
      setSessions(prev => prev.filter(s => s.id !== sessionId));
      if (currentSession?.id === sessionId) {
        const remainingSessions = sessions.filter(s => s.id !== sessionId);
        setCurrentSession(remainingSessions.length > 0 ? remainingSessions[0] : null);
      }
    } catch (error) {
      console.error('Failed to delete session:', error);
    }
  };

  const sendMessage = async () => {
    if (!message.trim() || isLoading || !currentSession) return;

    const userMessage: ChatMessage = {
      id: `msg_${Date.now()}_user`,
      role: 'user',
      content: message.trim(),
      timestamp: new Date(),
      metadata: {
        provider: selectedProvider,
        model: selectedModel
      }
    };

    // 添加用户消息到当前会话
    const updatedSession = {
      ...currentSession,
      messages: [...currentSession.messages, userMessage],
      updatedAt: new Date()
    };
    setCurrentSession(updatedSession);
    setSessions(prev => prev.map(s => s.id === currentSession.id ? updatedSession : s));

    const currentMessage = message;
    setMessage('');
    setIsLoading(true);

    try {
      // 使用新的AI服务发送消息
      const response = await aiService.sendChatMessage(currentMessage, currentSession.id, {
        model: selectedModel,
        provider: selectedProvider,
        temperature: 0.7,
        maxTokens: 2048
      });

      if (response.success) {
        const assistantMessage: ChatMessage = {
          id: `msg_${Date.now()}_assistant`,
          role: 'assistant',
          content: response.result.content,
          timestamp: new Date(),
          metadata: {
            provider: response.provider,
            model: response.model,
            tokens: response.result.tokensUsed,
            confidence: response.confidence
          }
        };

        // 添加AI回复到会话
        const finalSession = {
          ...updatedSession,
          messages: [...updatedSession.messages, assistantMessage],
          updatedAt: new Date()
        };
        setCurrentSession(finalSession);
        setSessions(prev => prev.map(s => s.id === currentSession.id ? finalSession : s));

        // 更新会话标题（如果是第一条消息）
        if (updatedSession.messages.length === 1) {
          const title = currentMessage.length > 30 ? currentMessage.substring(0, 30) + '...' : currentMessage;
          const sessionWithTitle = { ...finalSession, title };
          setCurrentSession(sessionWithTitle);
          setSessions(prev => prev.map(s => s.id === currentSession.id ? sessionWithTitle : s));
        }
      } else {
        // 处理错误
        const errorMessage: ChatMessage = {
          id: `msg_${Date.now()}_error`,
          role: 'assistant',
          content: `抱歉，发生了错误：${response.error}`,
          timestamp: new Date(),
          metadata: {
            provider: selectedProvider,
            model: selectedModel,
            error: response.error || 'Unknown error'
          }
        };

        const errorSession = {
          ...updatedSession,
          messages: [...updatedSession.messages, errorMessage],
          updatedAt: new Date()
        };
        setCurrentSession(errorSession);
        setSessions(prev => prev.map(s => s.id === currentSession.id ? errorSession : s));
      }
    } catch (error) {
      console.error('Failed to send message:', error);
      
      const errorMessage: ChatMessage = {
        id: `msg_${Date.now()}_error`,
        role: 'assistant',
        content: '抱歉，无法连接到AI服务，请稍后重试。',
        timestamp: new Date(),
        metadata: {
          provider: selectedProvider,
          model: selectedModel,
          error: 'Connection failed'
        }
      };

      const errorSession = {
        ...updatedSession,
        messages: [...updatedSession.messages, errorMessage],
        updatedAt: new Date()
      };
      setCurrentSession(errorSession);
      setSessions(prev => prev.map(s => s.id === currentSession.id ? errorSession : s));
    } finally {
      setIsLoading(false);
    }
  };



  const formatTime = (date: Date) => {
    return date.toLocaleTimeString('zh-CN', { 
      hour: '2-digit', 
      minute: '2-digit' 
    });
  };

  const getAvailableProviders = () => {
    return aiService.getAvailableProviders();
  };

  const getModelsForProvider = (providerId: string) => {
    return aiService.getModelsForProvider(providerId);
  };

  return (
    <div className="flex relative bg-gradient-to-br from-blue-50/30 via-purple-50/20 to-pink-50/30 dark:from-gray-900/50 dark:via-blue-900/20 dark:to-purple-900/30" style={{ height: '90vh' }}>
      {/* 移动端遮罩 */}
      {showSidebar && (
        <div 
          className="fixed inset-0 z-40 bg-black/60 backdrop-blur-sm lg:hidden fade-in"
          onClick={() => setShowSidebar(false)}
        />
      )}
      
      {/* Chat sessions sidebar */}
      <div className={cn(
        "w-80 border-r border-border/50 bg-white/80 dark:bg-gray-900/80 backdrop-blur-xl flex flex-col shadow-xl",
        "fixed inset-y-0 left-0 z-50 transform transition-all duration-300 ease-in-out lg:translate-x-0 lg:static lg:inset-0",
        showSidebar ? "translate-x-0 slide-in-left" : "-translate-x-full lg:translate-x-0"
      )}>
        <div className="p-4 border-b border-border/50 bg-gradient-to-r from-blue-50/50 to-purple-50/50 dark:from-gray-800/50 dark:to-blue-800/50">
          {/* 移动端关闭按钮 */}
          <div className="lg:hidden flex justify-end mb-4">
            <button
              onClick={() => setShowSidebar(false)}
              className="p-2 rounded-lg text-gray-600 hover:text-gray-900 hover:bg-white/80 dark:hover:bg-gray-700/80 transition-all duration-200 hover-lift"
            >
              <X className="h-5 w-5" />
            </button>
          </div>
          
          <button
            onClick={createNewSession}
            className="btn-primary w-full flex items-center justify-center mb-3 hover-lift bg-gradient-to-r from-blue-500 to-purple-600 hover:from-blue-600 hover:to-purple-700 shadow-lg hover:shadow-xl transition-all duration-200"
          >
            <Plus className="h-4 w-4 mr-2" />
            新建对话
          </button>
          <button
            onClick={() => setShowSettings(!showSettings)}
            className="btn-secondary w-full flex items-center justify-center hover-lift bg-gradient-to-r from-gray-100 to-gray-200 dark:from-gray-700 dark:to-gray-600 hover:from-gray-200 hover:to-gray-300 dark:hover:from-gray-600 dark:hover:to-gray-500 shadow-md hover:shadow-lg transition-all duration-200"
          >
            <Settings className="h-4 w-4 mr-2" />
            AI设置
          </button>
        </div>
        
        {showSettings && (
          <div className="p-4 border-b border-border/50 bg-gradient-to-r from-purple-50/50 to-pink-50/50 dark:from-gray-800/50 dark:to-purple-800/50 slide-in-down">
            <div className="space-y-4">
              <div className="scale-in">
                <label className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2 block">AI提供商</label>
                <select
                  value={selectedProvider}
                  onChange={(e) => setSelectedProvider(e.target.value)}
                  className="input w-full bg-white/80 dark:bg-gray-700/80 backdrop-blur-sm border-gray-200/50 dark:border-gray-600/50 focus:border-blue-400 focus:ring-2 focus:ring-blue-400/20 transition-all duration-200"
                >
                  {getAvailableProviders().map(provider => (
                    <option key={provider.id} value={provider.id}>
                      {provider.name}
                    </option>
                  ))}
                </select>
              </div>
              <div className="scale-in" style={{ animationDelay: '0.1s' }}>
                <label className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2 block">AI模型</label>
                <select
                  value={selectedModel}
                  onChange={(e) => setSelectedModel(e.target.value)}
                  className="input w-full bg-white/80 dark:bg-gray-700/80 backdrop-blur-sm border-gray-200/50 dark:border-gray-600/50 focus:border-blue-400 focus:ring-2 focus:ring-blue-400/20 transition-all duration-200"
                >
                  {getModelsForProvider(selectedProvider).map(model => (
                    <option key={model.id} value={model.id}>
                      {model.name}
                    </option>
                  ))}
                </select>
              </div>
            </div>
          </div>
        )}
        
        <div className="flex-1 overflow-y-auto p-4 space-y-3">
          {sessions.map((session, index) => (
            <div
              key={session.id}
              className={cn(
                'p-4 rounded-xl cursor-pointer transition-all duration-200 group hover-lift backdrop-blur-sm border border-white/20 dark:border-gray-700/30',
                'stagger-animation',
                currentSession?.id === session.id
                  ? 'bg-gradient-to-r from-blue-500/90 to-purple-600/90 text-white shadow-lg scale-105'
                  : 'bg-white/60 dark:bg-gray-800/60 hover:bg-white/80 dark:hover:bg-gray-700/80 hover:shadow-md'
              )}
              style={{ animationDelay: `${index * 0.1}s` }}
              onClick={() => setCurrentSession(session)}
            >
              <div className="flex items-center justify-between">
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-semibold truncate mb-1">{session.title}</p>
                  <p className={cn(
                    "text-xs flex items-center space-x-2",
                    currentSession?.id === session.id ? "text-white/80" : "text-gray-500 dark:text-gray-400"
                  )}>
                    <span>{session.messages.length} 条消息</span>
                    <span>•</span>
                    <span>{session.updatedAt.toLocaleDateString()}</span>
                  </p>
                </div>
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    deleteSession(session.id);
                  }}
                  className={cn(
                    "opacity-0 group-hover:opacity-100 p-2 rounded-lg transition-all duration-200 hover-lift",
                    currentSession?.id === session.id
                      ? "hover:bg-white/20 text-white/80 hover:text-white"
                      : "hover:bg-red-100 dark:hover:bg-red-900/30 text-gray-400 hover:text-red-600 dark:hover:text-red-400"
                  )}
                >
                  <Trash2 className="h-4 w-4" />
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Main chat area */}
      <div className="flex-1 flex flex-col lg:ml-0 bg-white/40 dark:bg-gray-900/40 backdrop-blur-sm">
        {/* 移动端顶部栏 */}
        <div className="lg:hidden bg-white/80 dark:bg-gray-900/80 backdrop-blur-xl border-b border-gray-200/50 dark:border-gray-700/50 px-4 py-3 shadow-sm">
          <div className="flex items-center justify-between">
            <button
              onClick={() => setShowSidebar(true)}
              className="p-2 rounded-lg text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white hover:bg-gray-100/80 dark:hover:bg-gray-700/80 transition-all duration-200 hover-lift"
            >
              <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
              </svg>
            </button>
            <h1 className="text-lg font-semibold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">AI对话</h1>
            <div className="w-10"></div>
          </div>
        </div>
        
        {/* Header */}
        <div className="p-4 border-b border-border/50 bg-gradient-to-r from-white/60 to-blue-50/60 dark:from-gray-900/60 dark:to-blue-900/40 backdrop-blur-sm">
          <div className="flex items-center justify-between">
            <div className="fade-in">
              <h2 className="text-lg font-semibold text-gray-800 dark:text-gray-200">
                {currentSession?.title || '选择或创建对话'}
              </h2>
              <p className="text-sm text-gray-600 dark:text-gray-400 flex items-center space-x-2">
                <span className="inline-block w-2 h-2 bg-green-400 rounded-full pulse-glow"></span>
                <span>{selectedProvider} - {selectedModel}</span>
              </p>
            </div>
            {currentSession && (
              <div className="text-sm text-gray-600 dark:text-gray-400 bg-white/60 dark:bg-gray-800/60 px-3 py-1 rounded-full backdrop-blur-sm border border-gray-200/50 dark:border-gray-700/50 scale-in">
                {currentSession.messages.length} 条消息
              </div>
            )}
          </div>
        </div>

        {/* Messages */}
        <div className="flex-1 overflow-y-auto p-6 space-y-6 bg-gradient-to-b from-transparent via-white/20 to-transparent dark:via-gray-900/20">
          {!currentSession || currentSession.messages.length === 0 ? (
            <div className="flex items-center justify-center h-full text-gray-500 dark:text-gray-400">
              <div className="text-center fade-in">
                <div className="relative mb-6">
                  <Bot className="h-16 w-16 mx-auto text-blue-400 dark:text-blue-300" />
                  <div className="absolute inset-0 h-16 w-16 mx-auto bg-blue-400/20 rounded-full animate-ping"></div>
                </div>
                <p className="text-lg font-medium mb-2 text-gray-700 dark:text-gray-300">开始与AI助手对话吧！</p>
                <p className="text-sm text-gray-500 dark:text-gray-400">发送消息开始对话</p>
              </div>
            </div>
          ) : (
            currentSession.messages.map((msg, index) => (
              <div
                key={msg.id}
                className={cn(
                  'flex items-start space-x-4 fade-in',
                  msg.role === 'user' ? 'justify-end' : 'justify-start'
                )}
                style={{ animationDelay: `${index * 0.1}s` }}
              >
                {msg.role === 'assistant' && (
                  <div className="h-10 w-10 rounded-full bg-gradient-to-r from-blue-500 to-purple-600 flex items-center justify-center flex-shrink-0 shadow-lg hover-lift">
                    <Bot className="h-5 w-5 text-white" />
                  </div>
                )}
                
                <div
                  className={cn(
                    'max-w-[75%] p-4 rounded-2xl shadow-md backdrop-blur-sm border hover-lift transition-all duration-200',
                    msg.role === 'user'
                      ? 'bg-gradient-to-r from-blue-500 to-purple-600 text-white border-blue-300/30 shadow-blue-200/50 dark:shadow-blue-900/30'
                      : msg.metadata?.error
                      ? 'bg-gradient-to-r from-red-500 to-pink-600 text-white border-red-300/30 shadow-red-200/50 dark:shadow-red-900/30'
                      : 'bg-white/80 dark:bg-gray-800/80 text-gray-800 dark:text-gray-200 border-gray-200/50 dark:border-gray-700/50 shadow-gray-200/50 dark:shadow-gray-900/30'
                  )}
                >
                  <p className="whitespace-pre-wrap leading-relaxed">{msg.content}</p>
                  <div className="flex items-center justify-between mt-3 pt-2 border-t border-white/20 dark:border-gray-700/30">
                    <p className={cn(
                      "text-xs flex items-center space-x-2",
                      msg.role === 'user' ? "text-white/80" : "text-gray-500 dark:text-gray-400"
                    )}>
                      <span>{formatTime(msg.timestamp)}</span>
                    </p>
                    {msg.metadata?.tokens && (
                      <p className={cn(
                        "text-xs px-2 py-1 rounded-full",
                        msg.role === 'user' 
                          ? "bg-white/20 text-white/80" 
                          : "bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300"
                      )}>
                        {msg.metadata.tokens} tokens
                      </p>
                    )}
                  </div>
                </div>

                {msg.role === 'user' && (
                  <div className="h-10 w-10 rounded-full bg-gradient-to-r from-green-500 to-teal-600 flex items-center justify-center flex-shrink-0 shadow-lg hover-lift">
                    <User className="h-5 w-5 text-white" />
                  </div>
                )}
              </div>
            ))
          )}
          
          {isLoading && (
            <div className="flex items-start space-x-4 fade-in">
              <div className="h-10 w-10 rounded-full bg-gradient-to-r from-blue-500 to-purple-600 flex items-center justify-center flex-shrink-0 shadow-lg">
                <Bot className="h-5 w-5 text-white" />
              </div>
              <div className="bg-white/80 dark:bg-gray-800/80 text-gray-800 dark:text-gray-200 p-4 rounded-2xl shadow-md backdrop-blur-sm border border-gray-200/50 dark:border-gray-700/50 hover-lift">
                <div className="flex items-center space-x-3">
                  <Loader2 className="h-5 w-5 animate-spin text-blue-500" />
                  <span className="text-sm">AI正在思考...</span>
                  <div className="flex space-x-1">
                    <div className="w-2 h-2 bg-blue-400 rounded-full animate-bounce"></div>
                    <div className="w-2 h-2 bg-purple-400 rounded-full animate-bounce" style={{ animationDelay: '0.1s' }}></div>
                    <div className="w-2 h-2 bg-pink-400 rounded-full animate-bounce" style={{ animationDelay: '0.2s' }}></div>
                  </div>
                </div>
              </div>
            </div>
          )}
          
          <div ref={messagesEndRef} />
        </div>

        {/* Input area */}
        <div className="p-4 border-t border-border/50 bg-gradient-to-r from-white/60 to-purple-50/60 dark:from-gray-900/60 dark:to-purple-900/40 backdrop-blur-xl">
          <div className="max-w-4xl mx-auto">
            <InputOptimizer
              value={message}
              onChange={setMessage}
              onSend={sendMessage}
              disabled={!currentSession}
              loading={isLoading}
              placeholder={currentSession ? "输入消息..." : "请先创建或选择对话"}
            />
          </div>
        </div>
      </div>
    </div>
  );
}