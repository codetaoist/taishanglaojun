/**
 * 聊天数据仓库
 * 整合本地数据存储和AI服务调用
 */

import relationalStore from '@ohos.data.relationalStore';
import { 
  ChatMessage, 
  Conversation, 
  MessageType, 
  MessageSender, 
  MessageStatus,
  AIPersonality,
  SendMessageRequest,
  CreateConversationRequest,
  ChatMessageEntity,
  ConversationEntity,
  ChatUtils
} from '../model/ChatMessage';
import { AIService, AIServiceInterface, MockAIService } from './AIService';
import { DataService } from './DataService';

// 聊天仓库接口
export interface ChatRepositoryInterface {
  // 对话管理
  createConversation(title: string, aiPersonality: AIPersonality): Promise<Conversation>;
  getConversations(): Promise<Conversation[]>;
  getConversation(conversationId: string): Promise<Conversation | null>;
  updateConversation(conversation: Conversation): Promise<void>;
  deleteConversation(conversationId: string): Promise<void>;
  archiveConversation(conversationId: string): Promise<void>;

  // 消息管理
  sendMessage(conversationId: string, content: string, messageType: MessageType): Promise<ChatMessage>;
  getMessages(conversationId: string, limit?: number, offset?: number): Promise<ChatMessage[]>;
  retryMessage(messageId: string): Promise<ChatMessage>;
  deleteMessage(messageId: string): Promise<void>;
  updateMessageStatus(messageId: string, status: MessageStatus): Promise<void>;

  // 搜索和统计
  searchMessages(query: string, conversationId?: string): Promise<ChatMessage[]>;
  getMessageCount(conversationId: string): Promise<number>;
  getUnreadCount(): Promise<number>;

  // 数据同步
  syncWithServer(): Promise<void>;
  clearLocalData(): Promise<void>;
}

// 聊天数据仓库实现
export class ChatRepository implements ChatRepositoryInterface {
  private static instance: ChatRepository;
  private aiService: AIServiceInterface;
  private dataService: DataService;
  private isInitialized: boolean = false;

  private constructor() {
    this.aiService = AIService.getInstance();
    this.dataService = DataService.getInstance();
  }

  static getInstance(): ChatRepository {
    if (!ChatRepository.instance) {
      ChatRepository.instance = new ChatRepository();
    }
    return ChatRepository.instance;
  }

  // 初始化
  async initialize(): Promise<void> {
    if (this.isInitialized) {
      return;
    }

    try {
      await this.dataService.initialize();
      this.isInitialized = true;
      console.info('ChatRepository initialized successfully');
    } catch (error) {
      console.error('Failed to initialize ChatRepository:', error);
      throw error;
    }
  }

  // 创建对话
  async createConversation(title: string, aiPersonality: AIPersonality): Promise<Conversation> {
    try {
      // 创建本地对话记录
      const conversationId = ChatUtils.generateId();
      const now = Date.now();
      
      const conversation: Conversation = {
        conversationId,
        title,
        aiPersonality,
        createdAt: now,
        updatedAt: now,
        messageCount: 0,
        isArchived: false
      };

      // 保存到本地数据库
      await this.saveConversationToLocal(conversation);

      // 尝试同步到服务器
      try {
        const request: CreateConversationRequest = {
          title,
          aiPersonality
        };
        await this.aiService.createConversation(request);
      } catch (error) {
        console.warn('Failed to sync conversation to server:', error);
        // 本地创建成功，服务器同步失败不影响使用
      }

      return conversation;
    } catch (error) {
      console.error('Failed to create conversation:', error);
      throw error;
    }
  }

  // 获取对话列表
  async getConversations(): Promise<Conversation[]> {
    try {
      const store = await this.dataService.getStore();
      const predicates = new relationalStore.RdbPredicates('conversations');
      predicates.orderByDesc('updatedAt');
      
      const resultSet = await store.query(predicates);
      const conversations: Conversation[] = [];

      if (resultSet.goToFirstRow()) {
        do {
          const entity = this.parseConversationFromResultSet(resultSet);
          conversations.push(this.entityToConversation(entity));
        } while (resultSet.goToNextRow());
      }

      resultSet.close();
      return conversations;
    } catch (error) {
      console.error('Failed to get conversations:', error);
      throw error;
    }
  }

  // 获取单个对话
  async getConversation(conversationId: string): Promise<Conversation | null> {
    try {
      const store = await this.dataService.getStore();
      const predicates = new relationalStore.RdbPredicates('conversations');
      predicates.equalTo('conversationId', conversationId);
      
      const resultSet = await store.query(predicates);
      let conversation: Conversation | null = null;

      if (resultSet.goToFirstRow()) {
        const entity = this.parseConversationFromResultSet(resultSet);
        conversation = this.entityToConversation(entity);
      }

      resultSet.close();
      return conversation;
    } catch (error) {
      console.error('Failed to get conversation:', error);
      throw error;
    }
  }

  // 更新对话
  async updateConversation(conversation: Conversation): Promise<void> {
    try {
      const store = await this.dataService.getStore();
      const entity = this.conversationToEntity(conversation);
      entity.updatedAt = Date.now();
      
      const predicates = new relationalStore.RdbPredicates('conversations');
      predicates.equalTo('conversationId', conversation.conversationId);
      
      await store.update(entity, predicates);
    } catch (error) {
      console.error('Failed to update conversation:', error);
      throw error;
    }
  }

  // 删除对话
  async deleteConversation(conversationId: string): Promise<void> {
    try {
      const store = await this.dataService.getStore();
      
      // 删除对话中的所有消息
      const messagePredicates = new relationalStore.RdbPredicates('chat_messages');
      messagePredicates.equalTo('conversationId', conversationId);
      await store.delete(messagePredicates);
      
      // 删除对话
      const conversationPredicates = new relationalStore.RdbPredicates('conversations');
      conversationPredicates.equalTo('conversationId', conversationId);
      await store.delete(conversationPredicates);

      // 尝试从服务器删除
      try {
        await this.aiService.deleteConversation(conversationId);
      } catch (error) {
        console.warn('Failed to delete conversation from server:', error);
      }
    } catch (error) {
      console.error('Failed to delete conversation:', error);
      throw error;
    }
  }

  // 归档对话
  async archiveConversation(conversationId: string): Promise<void> {
    try {
      const conversation = await this.getConversation(conversationId);
      if (conversation) {
        conversation.isArchived = true;
        await this.updateConversation(conversation);
      }
    } catch (error) {
      console.error('Failed to archive conversation:', error);
      throw error;
    }
  }

  // 发送消息
  async sendMessage(
    conversationId: string, 
    content: string, 
    messageType: MessageType = MessageType.TEXT,
    metadata?: Record<string, any>
  ): Promise<ChatMessage> {
    try {
      // 创建用户消息
      const userMessage: ChatMessage = {
        id: ChatUtils.generateId(),
        conversationId,
        content,
        messageType,
        sender: MessageSender.USER,
        status: MessageStatus.SENT,
        timestamp: Date.now(),
        metadata
      };

      // 保存用户消息到本地
      await this.saveMessageToLocal(userMessage);

      // 创建AI消息占位符
      const aiMessage: ChatMessage = {
        id: ChatUtils.generateId(),
        conversationId,
        content: '',
        messageType: MessageType.TEXT,
        sender: MessageSender.AI,
        status: MessageStatus.SENDING,
        timestamp: Date.now()
      };

      // 保存AI消息占位符
      await this.saveMessageToLocal(aiMessage);

      // 获取对话信息
      const conversation = await this.getConversation(conversationId);
      if (!conversation) {
        throw new Error('Conversation not found');
      }

      try {
        // 发送到AI服务
        const request: SendMessageRequest = {
          conversationId,
          message: content,
          messageType,
          aiPersonality: conversation.aiPersonality,
          metadata
        };

        const response = await this.aiService.sendMessage(request);

        if (response.success && response.message) {
          // 更新AI消息
          aiMessage.content = response.message.content;
          aiMessage.status = MessageStatus.SENT;
          aiMessage.metadata = response.message.metadata;
          
          await this.updateMessageInLocal(aiMessage);
        } else {
          // 处理错误
          aiMessage.status = MessageStatus.FAILED;
          aiMessage.error = response.error?.message || 'Unknown error';
          await this.updateMessageInLocal(aiMessage);
        }
      } catch (error) {
        console.error('Failed to send message to AI service:', error);
        aiMessage.status = MessageStatus.FAILED;
        aiMessage.error = error instanceof Error ? error.message : 'Network error';
        await this.updateMessageInLocal(aiMessage);
      }

      // 更新对话的最后更新时间和消息数量
      conversation.updatedAt = Date.now();
      conversation.messageCount = await this.getMessageCount(conversationId);
      await this.updateConversation(conversation);

      return aiMessage;
    } catch (error) {
      console.error('Failed to send message:', error);
      throw error;
    }
  }

  // 获取消息列表
  async getMessages(
    conversationId: string, 
    limit: number = 50, 
    offset: number = 0
  ): Promise<ChatMessage[]> {
    try {
      const store = await this.dataService.getStore();
      const predicates = new relationalStore.RdbPredicates('chat_messages');
      predicates.equalTo('conversationId', conversationId)
        .orderByDesc('timestamp')
        .limit(limit, offset);
      
      const resultSet = await store.query(predicates);
      const messages: ChatMessage[] = [];

      if (resultSet.goToFirstRow()) {
        do {
          const entity = this.parseMessageFromResultSet(resultSet);
          messages.push(this.entityToMessage(entity));
        } while (resultSet.goToNextRow());
      }

      resultSet.close();
      return messages.reverse(); // 按时间正序返回
    } catch (error) {
      console.error('Failed to get messages:', error);
      throw error;
    }
  }

  // 重试消息
  async retryMessage(messageId: string): Promise<ChatMessage> {
    try {
      const store = await this.dataService.getStore();
      const predicates = new relationalStore.RdbPredicates('chat_messages');
      predicates.equalTo('messageId', messageId);
      
      const resultSet = await store.query(predicates);
      if (!resultSet.goToFirstRow()) {
        resultSet.close();
        throw new Error('Message not found');
      }

      const entity = this.parseMessageFromResultSet(resultSet);
      const message = this.entityToMessage(entity);
      resultSet.close();

      if (message.sender !== MessageSender.AI || message.status !== MessageStatus.FAILED) {
        throw new Error('Only failed AI messages can be retried');
      }

      // 重新发送消息
      const conversation = await this.getConversation(message.conversationId);
      if (!conversation) {
        throw new Error('Conversation not found');
      }

      // 获取用户的原始消息
      const userMessages = await this.getMessages(message.conversationId);
      const userMessage = userMessages.find(m => 
        m.sender === MessageSender.USER && 
        m.timestamp < message.timestamp
      );

      if (!userMessage) {
        throw new Error('Original user message not found');
      }

      // 更新状态为发送中
      message.status = MessageStatus.SENDING;
      message.error = undefined;
      await this.updateMessageInLocal(message);

      try {
        const request: SendMessageRequest = {
          conversationId: message.conversationId,
          message: userMessage.content,
          messageType: userMessage.messageType,
          aiPersonality: conversation.aiPersonality
        };

        const response = await this.aiService.sendMessage(request);

        if (response.success && response.message) {
          message.content = response.message.content;
          message.status = MessageStatus.SENT;
          message.metadata = response.message.metadata;
        } else {
          message.status = MessageStatus.FAILED;
          message.error = response.error?.message || 'Unknown error';
        }
      } catch (error) {
        message.status = MessageStatus.FAILED;
        message.error = error instanceof Error ? error.message : 'Network error';
      }

      await this.updateMessageInLocal(message);
      return message;
    } catch (error) {
      console.error('Failed to retry message:', error);
      throw error;
    }
  }

  // 删除消息
  async deleteMessage(messageId: string): Promise<void> {
    try {
      const store = await this.dataService.getStore();
      const predicates = new relationalStore.RdbPredicates('chat_messages');
      predicates.equalTo('messageId', messageId);
      
      await store.delete(predicates);
    } catch (error) {
      console.error('Failed to delete message:', error);
      throw error;
    }
  }

  // 更新消息状态
  async updateMessageStatus(messageId: string, status: MessageStatus): Promise<void> {
    try {
      const store = await this.dataService.getStore();
      const predicates = new relationalStore.RdbPredicates('chat_messages');
      predicates.equalTo('messageId', messageId);
      
      const updateData = { status: status };
      await store.update(updateData, predicates);
    } catch (error) {
      console.error('Failed to update message status:', error);
      throw error;
    }
  }

  // 搜索消息
  async searchMessages(query: string, conversationId?: string): Promise<ChatMessage[]> {
    try {
      const store = await this.dataService.getStore();
      const predicates = new relationalStore.RdbPredicates('chat_messages');
      predicates.like('content', `%${query}%`);
      
      if (conversationId) {
        predicates.and().equalTo('conversationId', conversationId);
      }
      
      predicates.orderByDesc('timestamp');
      
      const resultSet = await store.query(predicates);
      const messages: ChatMessage[] = [];

      if (resultSet.goToFirstRow()) {
        do {
          const entity = this.parseMessageFromResultSet(resultSet);
          messages.push(this.entityToMessage(entity));
        } while (resultSet.goToNextRow());
      }

      resultSet.close();
      return messages;
    } catch (error) {
      console.error('Failed to search messages:', error);
      throw error;
    }
  }

  // 获取消息数量
  async getMessageCount(conversationId: string): Promise<number> {
    try {
      const store = await this.dataService.getStore();
      const predicates = new relationalStore.RdbPredicates('chat_messages');
      predicates.equalTo('conversationId', conversationId);
      
      const resultSet = await store.query(predicates);
      const count = resultSet.rowCount;
      resultSet.close();
      
      return count;
    } catch (error) {
      console.error('Failed to get message count:', error);
      return 0;
    }
  }

  // 获取未读消息数量
  async getUnreadCount(): Promise<number> {
    try {
      const store = await this.dataService.getStore();
      const predicates = new relationalStore.RdbPredicates('chat_messages');
      predicates.equalTo('isRead', false);
      
      const resultSet = await store.query(predicates);
      const count = resultSet.rowCount;
      resultSet.close();
      
      return count;
    } catch (error) {
      console.error('Failed to get unread count:', error);
      return 0;
    }
  }

  // 与服务器同步
  async syncWithServer(): Promise<void> {
    try {
      // 获取服务器上的对话列表
      const serverConversations = await this.aiService.getConversations();
      
      // 同步对话
      for (const conversation of serverConversations) {
        const localConversation = await this.getConversation(conversation.conversationId);
        if (!localConversation) {
          await this.saveConversationToLocal(conversation);
        }
      }

      console.info('Data synced with server successfully');
    } catch (error) {
      console.error('Failed to sync with server:', error);
      throw error;
    }
  }

  // 清除本地数据
  async clearLocalData(): Promise<void> {
    try {
      const store = await this.dataService.getStore();
      
      // 清除所有消息
      await store.delete(new relationalStore.RdbPredicates('chat_messages'));
      
      // 清除所有对话
      await store.delete(new relationalStore.RdbPredicates('conversations'));
      
      console.info('Local data cleared successfully');
    } catch (error) {
      console.error('Failed to clear local data:', error);
      throw error;
    }
  }

  // 私有方法：保存对话到本地
  private async saveConversationToLocal(conversation: Conversation): Promise<void> {
    const store = await this.dataService.getStore();
    const entity = this.conversationToEntity(conversation);
    await store.insert('conversations', entity);
  }

  // 私有方法：保存消息到本地
  private async saveMessageToLocal(message: ChatMessage): Promise<void> {
    const store = await this.dataService.getStore();
    const entity = this.messageToEntity(message);
    await store.insert('chat_messages', entity);
  }

  // 私有方法：更新本地消息
  private async updateMessageInLocal(message: ChatMessage): Promise<void> {
    const store = await this.dataService.getStore();
    const entity = this.messageToEntity(message);
    
    const predicates = new relationalStore.RdbPredicates('chat_messages');
    predicates.equalTo('messageId', message.messageId);
    
    await store.update(entity, predicates);
  }

  // 数据转换方法
  private conversationToEntity(conversation: Conversation): ConversationEntity {
    return {
      conversationId: conversation.conversationId,
      title: conversation.title,
      aiPersonality: conversation.aiPersonality,
      createdAt: conversation.createdAt,
      updatedAt: conversation.updatedAt,
      messageCount: conversation.messageCount,
      isArchived: conversation.isArchived ? 1 : 0
    };
  }

  private entityToConversation(entity: ConversationEntity): Conversation {
    return {
      conversationId: entity.conversationId,
      title: entity.title,
      aiPersonality: entity.aiPersonality,
      createdAt: entity.createdAt,
      updatedAt: entity.updatedAt,
      messageCount: entity.messageCount,
      isArchived: entity.isArchived === 1
    };
  }

  private messageToEntity(message: ChatMessage): ChatMessageEntity {
    return {
      messageId: message.messageId,
      conversationId: message.conversationId,
      content: message.content,
      messageType: message.messageType,
      sender: message.sender,
      status: message.status,
      timestamp: message.timestamp,
      isRead: message.isRead ? 1 : 0,
      metadata: message.metadata ? JSON.stringify(message.metadata) : null,
      error: message.error || null
    };
  }

  private entityToMessage(entity: ChatMessageEntity): ChatMessage {
    return {
      messageId: entity.messageId,
      conversationId: entity.conversationId,
      content: entity.content,
      messageType: entity.messageType,
      sender: entity.sender,
      status: entity.status,
      timestamp: entity.timestamp,
      isRead: entity.isRead === 1,
      metadata: entity.metadata ? JSON.parse(entity.metadata) : undefined,
      error: entity.error || undefined
    };
  }

  // 解析数据库结果集
  private parseConversationFromResultSet(resultSet: relationalStore.ResultSet): ConversationEntity {
    return {
      conversationId: resultSet.getString(resultSet.getColumnIndex('conversationId')),
      title: resultSet.getString(resultSet.getColumnIndex('title')),
      aiPersonality: resultSet.getString(resultSet.getColumnIndex('aiPersonality')) as AIPersonality,
      createdAt: resultSet.getLong(resultSet.getColumnIndex('createdAt')),
      updatedAt: resultSet.getLong(resultSet.getColumnIndex('updatedAt')),
      messageCount: resultSet.getLong(resultSet.getColumnIndex('messageCount')),
      isArchived: resultSet.getLong(resultSet.getColumnIndex('isArchived'))
    };
  }

  private parseMessageFromResultSet(resultSet: relationalStore.ResultSet): ChatMessageEntity {
    return {
      messageId: resultSet.getString(resultSet.getColumnIndex('messageId')),
      conversationId: resultSet.getString(resultSet.getColumnIndex('conversationId')),
      content: resultSet.getString(resultSet.getColumnIndex('content')),
      messageType: resultSet.getString(resultSet.getColumnIndex('messageType')) as MessageType,
      sender: resultSet.getString(resultSet.getColumnIndex('sender')) as MessageSender,
      status: resultSet.getString(resultSet.getColumnIndex('status')) as MessageStatus,
      timestamp: resultSet.getLong(resultSet.getColumnIndex('timestamp')),
      isRead: resultSet.getLong(resultSet.getColumnIndex('isRead')),
      metadata: resultSet.getString(resultSet.getColumnIndex('metadata')),
      error: resultSet.getString(resultSet.getColumnIndex('error'))
    };
  }
}