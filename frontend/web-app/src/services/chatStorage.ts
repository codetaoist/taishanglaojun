/**
 * 聊天存储服务
 * 负责管理对话历史记录的本地存储和会话持久化
 */

import type { ChatMessage, Conversation, ConversationSummary } from '../types';

class ChatStorageService {
  private readonly STORAGE_KEY = 'taishanglaojun_chat_conversations';
  private readonly CURRENT_SESSION_KEY = 'taishanglaojun_current_session';
  private readonly MAX_CONVERSATIONS = 50; // 最大保存对话数量
  private readonly MAX_MESSAGES_PER_CONVERSATION = 100; // 每个对话最大消息数量

  /**
   * 获取所有对话摘要
   */
  getConversationSummaries(): ConversationSummary[] {
    try {
      const stored = localStorage.getItem(this.STORAGE_KEY);
      if (!stored) return [];

      const conversations: Conversation[] = JSON.parse(stored);
      return conversations
        .filter(conv => !conv.isArchived)
        .map(conv => ({
          id: conv.id,
          title: conv.title,
          lastMessage: conv.messages.length > 0 
            ? conv.messages[conv.messages.length - 1].content.substring(0, 100) + '...'
            : '暂无消息',
          messageCount: conv.messageCount,
          createdAt: conv.createdAt,
          updatedAt: conv.updatedAt,
          isArchived: conv.isArchived
        }))
        .sort((a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime());
    } catch (error) {
      console.error('Failed to get conversation summaries:', error);
      return [];
    }
  }

  /**
   * 获取指定对话的完整信息
   */
  getConversation(conversationId: string): Conversation | null {
    try {
      const stored = localStorage.getItem(this.STORAGE_KEY);
      if (!stored) return null;

      const conversations: Conversation[] = JSON.parse(stored);
      return conversations.find(conv => conv.id === conversationId) || null;
    } catch (error) {
      console.error('Failed to get conversation:', error);
      return null;
    }
  }

  /**
   * 创建新对话
   */
  createConversation(title?: string): Conversation {
    const now = new Date().toISOString();
    const conversation: Conversation = {
      id: this.generateId(),
      title: title || `对话 ${new Date().toLocaleString()}`,
      messages: [],
      createdAt: now,
      updatedAt: now,
      isArchived: false,
      messageCount: 0
    };

    this.saveConversation(conversation);
    return conversation;
  }

  /**
   * 保存对话
   */
  saveConversation(conversation: Conversation): void {
    try {
      const stored = localStorage.getItem(this.STORAGE_KEY);
      let conversations: Conversation[] = stored ? JSON.parse(stored) : [];

      // 更新或添加对话
      const existingIndex = conversations.findIndex(conv => conv.id === conversation.id);
      if (existingIndex >= 0) {
        conversations[existingIndex] = conversation;
      } else {
        conversations.unshift(conversation);
      }

      // 限制对话数量
      if (conversations.length > this.MAX_CONVERSATIONS) {
        conversations = conversations.slice(0, this.MAX_CONVERSATIONS);
      }

      // 限制每个对话的消息数量
      conversations.forEach(conv => {
        if (conv.messages.length > this.MAX_MESSAGES_PER_CONVERSATION) {
          conv.messages = conv.messages.slice(-this.MAX_MESSAGES_PER_CONVERSATION);
          conv.messageCount = conv.messages.length;
        }
      });

      localStorage.setItem(this.STORAGE_KEY, JSON.stringify(conversations));
    } catch (error) {
      console.error('Failed to save conversation:', error);
    }
  }

  /**
   * 添加消息到对话
   */
  addMessage(conversationId: string, message: ChatMessage): void {
    const conversation = this.getConversation(conversationId);
    if (!conversation) return;

    conversation.messages.push(message);
    conversation.messageCount = conversation.messages.length;
    conversation.updatedAt = new Date().toISOString();

    // 自动生成对话标题（基于第一条用户消息）
    if (conversation.messages.length === 1 && message.role === 'user') {
      conversation.title = message.content.substring(0, 30) + (message.content.length > 30 ? '...' : '');
    }

    this.saveConversation(conversation);
  }

  /**
   * 更新消息
   */
  updateMessage(conversationId: string, messageId: string, updates: Partial<ChatMessage>): void {
    const conversation = this.getConversation(conversationId);
    if (!conversation) return;

    const messageIndex = conversation.messages.findIndex(msg => msg.id === messageId);
    if (messageIndex >= 0) {
      conversation.messages[messageIndex] = { ...conversation.messages[messageIndex], ...updates };
      conversation.updatedAt = new Date().toISOString();
      this.saveConversation(conversation);
    }
  }

  /**
   * 删除对话
   */
  deleteConversation(conversationId: string): void {
    try {
      const stored = localStorage.getItem(this.STORAGE_KEY);
      if (!stored) return;

      let conversations: Conversation[] = JSON.parse(stored);
      conversations = conversations.filter(conv => conv.id !== conversationId);
      localStorage.setItem(this.STORAGE_KEY, JSON.stringify(conversations));
    } catch (error) {
      console.error('Failed to delete conversation:', error);
    }
  }

  /**
   * 归档对话
   */
  archiveConversation(conversationId: string): void {
    const conversation = this.getConversation(conversationId);
    if (conversation) {
      conversation.isArchived = true;
      conversation.updatedAt = new Date().toISOString();
      this.saveConversation(conversation);
    }
  }

  /**
   * 获取当前会话信息
   */
  getCurrentSession(): { conversationId?: string; sessionId?: string } {
    try {
      const stored = sessionStorage.getItem(this.CURRENT_SESSION_KEY);
      return stored ? JSON.parse(stored) : {};
    } catch (error) {
      console.error('Failed to get current session:', error);
      return {};
    }
  }

  /**
   * 设置当前会话信息
   */
  setCurrentSession(conversationId?: string, sessionId?: string): void {
    try {
      const session = { conversationId, sessionId };
      sessionStorage.setItem(this.CURRENT_SESSION_KEY, JSON.stringify(session));
    } catch (error) {
      console.error('Failed to set current session:', error);
    }
  }

  /**
   * 清除当前会话
   */
  clearCurrentSession(): void {
    sessionStorage.removeItem(this.CURRENT_SESSION_KEY);
  }

  /**
   * 搜索对话
   */
  searchConversations(query: string): ConversationSummary[] {
    const summaries = this.getConversationSummaries();
    const lowerQuery = query.toLowerCase();
    
    return summaries.filter(summary => 
      summary.title.toLowerCase().includes(lowerQuery) ||
      summary.lastMessage.toLowerCase().includes(lowerQuery)
    );
  }

  /**
   * 搜索消息
   */
  searchMessages(conversationId: string, query: string): ChatMessage[] {
    const conversation = this.getConversation(conversationId);
    if (!conversation) return [];

    const lowerQuery = query.toLowerCase();
    return conversation.messages.filter(message =>
      message.content.toLowerCase().includes(lowerQuery)
    );
  }

  /**
   * 清理过期数据
   */
  cleanupOldData(daysToKeep: number = 30): void {
    try {
      const stored = localStorage.getItem(this.STORAGE_KEY);
      if (!stored) return;

      const conversations: Conversation[] = JSON.parse(stored);
      const cutoffDate = new Date();
      cutoffDate.setDate(cutoffDate.getDate() - daysToKeep);

      const filteredConversations = conversations.filter(conv => {
        const updatedAt = new Date(conv.updatedAt);
        return updatedAt > cutoffDate || !conv.isArchived;
      });

      localStorage.setItem(this.STORAGE_KEY, JSON.stringify(filteredConversations));
    } catch (error) {
      console.error('Failed to cleanup old data:', error);
    }
  }

  /**
   * 导出对话数据
   */
  exportConversations(): string {
    try {
      const stored = localStorage.getItem(this.STORAGE_KEY);
      return stored || '[]';
    } catch (error) {
      console.error('Failed to export conversations:', error);
      return '[]';
    }
  }

  /**
   * 导入对话数据
   */
  importConversations(data: string): boolean {
    try {
      const conversations: Conversation[] = JSON.parse(data);
      
      // 验证数据格式
      if (!Array.isArray(conversations)) {
        throw new Error('Invalid data format');
      }

      localStorage.setItem(this.STORAGE_KEY, JSON.stringify(conversations));
      return true;
    } catch (error) {
      console.error('Failed to import conversations:', error);
      return false;
    }
  }

  /**
   * 获取存储统计信息
   */
  getStorageStats(): {
    totalConversations: number;
    totalMessages: number;
    archivedConversations: number;
    storageSize: number;
  } {
    try {
      const stored = localStorage.getItem(this.STORAGE_KEY);
      if (!stored) {
        return {
          totalConversations: 0,
          totalMessages: 0,
          archivedConversations: 0,
          storageSize: 0
        };
      }

      const conversations: Conversation[] = JSON.parse(stored);
      const totalMessages = conversations.reduce((sum, conv) => sum + conv.messageCount, 0);
      const archivedConversations = conversations.filter(conv => conv.isArchived).length;
      const storageSize = new Blob([stored]).size;

      return {
        totalConversations: conversations.length,
        totalMessages,
        archivedConversations,
        storageSize
      };
    } catch (error) {
      console.error('Failed to get storage stats:', error);
      return {
        totalConversations: 0,
        totalMessages: 0,
        archivedConversations: 0,
        storageSize: 0
      };
    }
  }

  /**
   * 生成唯一ID
   */
  private generateId(): string {
    return `chat_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }
}

// 导出单例实例
export const chatStorage = new ChatStorageService();