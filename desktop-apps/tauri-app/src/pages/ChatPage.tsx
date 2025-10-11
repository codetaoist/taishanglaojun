import { useState, useRef, useEffect } from 'react';
import { Send, Bot, User, Loader2, Plus, Trash2, Settings } from 'lucide-react';
import { cn } from '../utils/cn';
import aiService, { ChatMessage, ChatSession } from '../services/aiService';

export default function ChatPage() {
  const [sessions, setSessions] = useState<ChatSession[]>([]);
  const [currentSession, setCurrentSession] = useState<ChatSession | null>(null);
  const [message, setMessage] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [selectedModel, setSelectedModel] = useState('taishanglaojun-chat');
  const [selectedProvider, setSelectedProvider] = useState('local');
  const [showSettings, setShowSettings] = useState(false);
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
            error: true
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
          error: true
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

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
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
    <div className="flex h-full">
      {/* Chat sessions sidebar */}
      <div className="w-80 border-r border-border bg-secondary/5 flex flex-col">
        <div className="p-4 border-b border-border">
          <button
            onClick={createNewSession}
            className="btn-primary w-full flex items-center justify-center mb-2"
          >
            <Plus className="h-4 w-4 mr-2" />
            新建对话
          </button>
          <button
            onClick={() => setShowSettings(!showSettings)}
            className="btn-secondary w-full flex items-center justify-center"
          >
            <Settings className="h-4 w-4 mr-2" />
            AI设置
          </button>
        </div>
        
        {showSettings && (
          <div className="p-4 border-b border-border bg-background/50">
            <div className="space-y-3">
              <div>
                <label className="text-sm font-medium">AI提供商</label>
                <select
                  value={selectedProvider}
                  onChange={(e) => setSelectedProvider(e.target.value)}
                  className="input w-full mt-1"
                >
                  {getAvailableProviders().map(provider => (
                    <option key={provider.id} value={provider.id}>
                      {provider.name}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="text-sm font-medium">AI模型</label>
                <select
                  value={selectedModel}
                  onChange={(e) => setSelectedModel(e.target.value)}
                  className="input w-full mt-1"
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
        
        <div className="flex-1 overflow-y-auto p-4 space-y-2">
          {sessions.map((session) => (
            <div
              key={session.id}
              className={cn(
                'p-3 rounded-lg cursor-pointer transition-colors group',
                currentSession?.id === session.id
                  ? 'bg-primary text-primary-foreground'
                  : 'hover:bg-accent hover:text-accent-foreground'
              )}
              onClick={() => setCurrentSession(session)}
            >
              <div className="flex items-center justify-between">
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium truncate">{session.title}</p>
                  <p className="text-xs opacity-70">
                    {session.messages.length} 条消息
                  </p>
                </div>
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    deleteSession(session.id);
                  }}
                  className="opacity-0 group-hover:opacity-100 p-1 hover:bg-destructive hover:text-destructive-foreground rounded transition-all"
                >
                  <Trash2 className="h-4 w-4" />
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Main chat area */}
      <div className="flex-1 flex flex-col">
        {/* Header */}
        <div className="p-4 border-b border-border bg-background/50">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-lg font-semibold">
                {currentSession?.title || '选择或创建对话'}
              </h2>
              <p className="text-sm text-muted-foreground">
                {selectedProvider} - {selectedModel}
              </p>
            </div>
            {currentSession && (
              <div className="text-sm text-muted-foreground">
                {currentSession.messages.length} 条消息
              </div>
            )}
          </div>
        </div>

        {/* Messages */}
        <div className="flex-1 overflow-y-auto p-4 space-y-4">
          {!currentSession || currentSession.messages.length === 0 ? (
            <div className="flex items-center justify-center h-full text-muted-foreground">
              <div className="text-center">
                <Bot className="h-12 w-12 mx-auto mb-4 opacity-50" />
                <p>开始与AI助手对话吧！</p>
                <p className="text-sm mt-2">发送消息开始对话</p>
              </div>
            </div>
          ) : (
            currentSession.messages.map((msg) => (
              <div
                key={msg.id}
                className={cn(
                  'flex items-start space-x-3',
                  msg.role === 'user' ? 'justify-end' : 'justify-start'
                )}
              >
                {msg.role === 'assistant' && (
                  <div className="h-8 w-8 rounded-full bg-primary flex items-center justify-center flex-shrink-0">
                    <Bot className="h-4 w-4 text-primary-foreground" />
                  </div>
                )}
                
                <div
                  className={cn(
                    'max-w-[70%] p-3 rounded-lg',
                    msg.role === 'user'
                      ? 'bg-primary text-primary-foreground'
                      : msg.metadata?.error
                      ? 'bg-destructive text-destructive-foreground'
                      : 'bg-secondary text-secondary-foreground'
                  )}
                >
                  <p className="whitespace-pre-wrap">{msg.content}</p>
                  <div className="flex items-center justify-between mt-2">
                    <p className="text-xs opacity-70">
                      {formatTime(msg.timestamp)}
                    </p>
                    {msg.metadata?.tokens && (
                      <p className="text-xs opacity-70">
                        {msg.metadata.tokens} tokens
                      </p>
                    )}
                  </div>
                </div>

                {msg.role === 'user' && (
                  <div className="h-8 w-8 rounded-full bg-secondary flex items-center justify-center flex-shrink-0">
                    <User className="h-4 w-4 text-secondary-foreground" />
                  </div>
                )}
              </div>
            ))
          )}
          
          {isLoading && (
            <div className="flex items-start space-x-3">
              <div className="h-8 w-8 rounded-full bg-primary flex items-center justify-center flex-shrink-0">
                <Bot className="h-4 w-4 text-primary-foreground" />
              </div>
              <div className="bg-secondary text-secondary-foreground p-3 rounded-lg">
                <div className="flex items-center space-x-2">
                  <Loader2 className="h-4 w-4 animate-spin" />
                  <span>AI正在思考...</span>
                </div>
              </div>
            </div>
          )}
          
          <div ref={messagesEndRef} />
        </div>

        {/* Input area */}
        <div className="p-4 border-t border-border bg-background/50">
          <div className="flex space-x-2">
            <textarea
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              onKeyPress={handleKeyPress}
              placeholder={currentSession ? "输入消息..." : "请先创建或选择对话"}
              className="input flex-1 min-h-[44px] max-h-32 resize-none"
              rows={1}
              disabled={!currentSession}
            />
            <button
              onClick={sendMessage}
              disabled={!message.trim() || isLoading || !currentSession}
              className={cn(
                'btn-primary px-4',
                (!message.trim() || isLoading || !currentSession) && 'opacity-50 cursor-not-allowed'
              )}
            >
              {isLoading ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <Send className="h-4 w-4" />
              )}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}