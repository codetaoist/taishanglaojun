/**
 * 聊天管理Hook
 * 提供聊天功能的状态管理和业务逻辑
 */

import { useState, useEffect, useCallback, useRef, useMemo } from 'react';
import { message } from 'antd';
import { apiClient } from '../services/api';
import { chatStorage } from '../services/chatStorage';
import type { ChatMessage, Conversation, ConversationSummary } from '../types';

export interface UseChatOptions {
  autoSave?: boolean;
  maxRetries?: number;
  retryDelay?: number;
}

export interface UseChatReturn {
  // 当前对话状态
  currentConversation: Conversation | null;
  messages: ChatMessage[];
  loading: boolean;
  
  // 对话列表
  conversations: ConversationSummary[];
  
  // 会话管理
  sessionId: string | undefined;
  
  // 操作方法
  sendMessage: (content: string) => Promise<void>;
  createNewConversation: (title?: string) => void;
  loadConversation: (conversationId: string) => void;
  deleteConversation: (conversationId: string) => void;
  archiveConversation: (conversationId: string) => void;
  clearCurrentConversation: () => void;
  retryLastMessage: () => Promise<void>;
  
  // 搜索功能
  searchConversations: (query: string) => ConversationSummary[];
  searchMessages: (query: string) => ChatMessage[];
  
  // 工具方法
  exportConversations: () => string;
  importConversations: (data: string) => boolean;
  getStorageStats: () => any;
}

export const useChat = (options: UseChatOptions = {}): UseChatReturn => {
  const {
    autoSave = true,
    maxRetries = 3,
    retryDelay = 1000
  } = options;

  // 状态管理
  const [currentConversation, setCurrentConversation] = useState<Conversation | null>(null);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [loading, setLoading] = useState(false);
  const [conversations, setConversations] = useState<ConversationSummary[]>([]);
  const [sessionId, setSessionId] = useState<string | undefined>();
  
  // 引用
  const retryCountRef = useRef<number>(0);
  const lastMessageRef = useRef<string>('');

  // 使用useMemo缓存搜索函数
  const searchConversations = useMemo(() => {
    return (query: string): ConversationSummary[] => {
      if (!query.trim()) return conversations;
      
      const searchLower = query.toLowerCase();
      return conversations.filter(conv => 
        conv.title.toLowerCase().includes(searchLower) ||
        conv.preview.toLowerCase().includes(searchLower)
      );
    };
  }, [conversations]);

  const searchMessages = useMemo(() => {
    return (query: string): ChatMessage[] => {
      if (!query.trim()) return messages;
      
      const searchLower = query.toLowerCase();
      return messages.filter(msg => 
        msg.content.toLowerCase().includes(searchLower)
      );
    };
  }, [messages]);

  // 使用useMemo缓存统计信息
  const storageStats = useMemo(() => {
    return {
      totalConversations: conversations.length,
      totalMessages: messages.length,
      currentConversationId: currentConversation?.id,
      sessionId,
      lastActivity: currentConversation?.updatedAt,
    };
  }, [conversations.length, messages.length, currentConversation?.id, currentConversation?.updatedAt, sessionId]);

  // 初始化
  useEffect(() => {
    loadConversationList();
    restoreCurrentSession();
  }, []);

  // 自动保存当前会话信息
  useEffect(() => {
    if (autoSave && currentConversation) {
      chatStorage.setCurrentSession(currentConversation.id, sessionId);
    }
  }, [currentConversation?.id, sessionId, autoSave]);

  /**
   * 加载对话列表
   */
  const loadConversationList = useCallback(() => {
    const summaries = chatStorage.getConversationSummaries();
    setConversations(summaries);
  }, []);

  /**
   * 恢复当前会话
   */
  const restoreCurrentSession = useCallback(() => {
    const session = chatStorage.getCurrentSession();
    if (session.conversationId) {
      const conversation = chatStorage.getConversation(session.conversationId);
      if (conversation) {
        setCurrentConversation(conversation);
        setMessages(conversation.messages);
        setSessionId(session.sessionId);
      }
    }
  }, []);

  /**
   * 发送消息
   */
  const sendMessage = useCallback(async (content: string) => {
    if (!content.trim() || loading) return;

    // 确保有当前对话
    let conversation = currentConversation;
    if (!conversation) {
      conversation = chatStorage.createConversation();
      setCurrentConversation(conversation);
    }

    // 创建用户消息
    const userMessage: ChatMessage = {
      id: `msg_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      content: content.trim(),
      role: 'user',
      timestamp: new Date().toISOString(),
    };

    // 立即显示用户消息
    const newMessages = [...messages, userMessage];
    setMessages(newMessages);
    lastMessageRef.current = content.trim();

    // 保存用户消息
    chatStorage.addMessage(conversation.id, userMessage);

    setLoading(true);
    retryCountRef.current = 0;

    try {
      await sendMessageWithRetry(content.trim(), conversation.id);
    } catch (error) {
      console.error('Failed to send message after retries:', error);
      message.error('发送消息失败，请检查网络连接');
      
      // 移除失败的用户消息
      setMessages(prev => prev.slice(0, -1));
    } finally {
      setLoading(false);
      loadConversationList(); // 刷新对话列表
    }
  }, [currentConversation, messages, loading, maxRetries, retryDelay]);

  /**
   * 带重试的发送消息
   */
  const sendMessageWithRetry = useCallback(async (content: string, conversationId: string): Promise<void> => {
    try {
      const response = await apiClient.sendChatMessage(content, sessionId);
      
      if (response.success) {
        // 创建AI回复消息
        const aiMessage: ChatMessage = {
          id: response.data.message_id.toString(),
          content: response.data.content,
          role: 'assistant',
          timestamp: new Date().toISOString(),
          metadata: {
            sources: ['AI助手'],
            confidence: 0.9
          }
        };
        
        // 更新消息列表
        setMessages(prev => [...prev, aiMessage]);
        
        // 保存AI消息
        chatStorage.addMessage(conversationId, aiMessage);
        
        // 更新会话ID
        if (!sessionId && response.data.session_id) {
          setSessionId(response.data.session_id);
        }
        
        // 重置重试计数
        retryCountRef.current = 0;
      } else {
        throw new Error(response.message || 'API调用失败');
      }
    } catch (error) {
      retryCountRef.current++;
      
      if (retryCountRef.current < maxRetries) {
        console.warn(`Message send failed, retrying (${retryCountRef.current}/${maxRetries}):`, error);
        await new Promise(resolve => setTimeout(resolve, retryDelay * retryCountRef.current));
        return sendMessageWithRetry(content, conversationId);
      } else {
        throw error;
      }
    }
  }, [sessionId, maxRetries, retryDelay]);

  /**
   * 重试最后一条消息
   */
  const retryLastMessage = useCallback(async () => {
    if (!lastMessageRef.current || loading) return;
    
    // 移除最后一条AI消息（如果存在且失败）
    setMessages(prev => {
      const lastMessage = prev[prev.length - 1];
      if (lastMessage && lastMessage.role === 'assistant') {
        return prev.slice(0, -1);
      }
      return prev;
    });
    
    await sendMessage(lastMessageRef.current);
  }, [sendMessage, loading]);

  /**
   * 创建新对话
   */
  const createNewConversation = useCallback((title?: string) => {
    const conversation = chatStorage.createConversation(title);
    setCurrentConversation(conversation);
    setMessages([]);
    setSessionId(undefined);
    chatStorage.clearCurrentSession();
    loadConversationList();
  }, [loadConversationList]);

  /**
   * 加载指定对话
   */
  const loadConversation = useCallback((conversationId: string) => {
    const conversation = chatStorage.getConversation(conversationId);
    if (conversation) {
      setCurrentConversation(conversation);
      setMessages(conversation.messages);
      setSessionId(conversation.sessionId);
      chatStorage.setCurrentSession(conversation.id, conversation.sessionId);
    }
  }, []);

  /**
   * 删除对话
   */
  const deleteConversation = useCallback((conversationId: string) => {
    chatStorage.deleteConversation(conversationId);
    
    // 如果删除的是当前对话，清空当前状态
    if (currentConversation?.id === conversationId) {
      setCurrentConversation(null);
      setMessages([]);
      setSessionId(undefined);
      chatStorage.clearCurrentSession();
    }
    
    loadConversationList();
  }, [currentConversation?.id, loadConversationList]);

  /**
   * 归档对话
   */
  const archiveConversation = useCallback((conversationId: string) => {
    chatStorage.archiveConversation(conversationId);
    
    // 如果归档的是当前对话，清空当前状态
    if (currentConversation?.id === conversationId) {
      setCurrentConversation(null);
      setMessages([]);
      setSessionId(undefined);
      chatStorage.clearCurrentSession();
    }
    
    loadConversationList();
  }, [currentConversation?.id, loadConversationList]);

  /**
   * 清空当前对话
   */
  const clearCurrentConversation = useCallback(() => {
    setCurrentConversation(null);
    setMessages([]);
    setSessionId(undefined);
    chatStorage.clearCurrentSession();
  }, []);



  /**
   * 导出对话
   */
  const exportConversations = useCallback((): string => {
    return chatStorage.exportConversations();
  }, []);

  /**
   * 导入对话
   */
  const importConversations = useCallback((data: string): boolean => {
    const success = chatStorage.importConversations(data);
    if (success) {
      loadConversationList();
      message.success('对话数据导入成功');
    } else {
      message.error('对话数据导入失败，请检查数据格式');
    }
    return success;
  }, [loadConversationList]);

  /**
   * 获取存储统计
   */
  const getStorageStats = useCallback(() => {
    return chatStorage.getStorageStats();
  }, []);

  return {
    // 状态
    currentConversation,
    messages,
    loading,
    conversations,
    sessionId,
    
    // 方法
    sendMessage,
    createNewConversation,
    loadConversation,
    deleteConversation,
    archiveConversation,
    clearCurrentConversation,
    retryLastMessage,
    searchConversations,
    searchMessages,
    exportConversations,
    importConversations,
    getStorageStats: () => storageStats,
  };
};