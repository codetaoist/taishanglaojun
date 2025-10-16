/**
 * AI对话消息数据模型
 * 支持鸿蒙应用的AI对话功能
 */

// 消息类型枚举
export enum MessageType {
  TEXT = 'text',
  IMAGE = 'image',
  VOICE = 'voice',
  FILE = 'file'
}

// 消息发送者枚举
export enum MessageSender {
  USER = 'user',
  AI = 'ai'
}

// 消息状态枚举
export enum MessageStatus {
  SENDING = 'sending',
  SENT = 'sent',
  FAILED = 'failed',
  RECEIVED = 'received'
}

// AI人格枚举
export enum AIPersonality {
  DEFAULT = 'default',
  WISE_SAGE = 'wise_sage',
  FRIENDLY_GUIDE = 'friendly_guide',
  SCHOLARLY = 'scholarly',
  POETIC = 'poetic'
}

// 聊天消息接口
export interface ChatMessage {
  id: string;
  conversationId: string;
  content: string;
  messageType: MessageType;
  sender: MessageSender;
  status: MessageStatus;
  timestamp: number;
  metadata?: Record<string, any>;
}

// 对话接口
export interface Conversation {
  id: string;
  title: string;
  aiPersonality: AIPersonality;
  createdAt: number;
  updatedAt: number;
  isArchived: boolean;
  messageCount?: number;
  lastMessage?: string;
}

// AI响应接口
export interface AIResponse {
  success: boolean;
  message?: AIMessage;
  suggestions?: string[];
  error?: AIError;
}

// AI消息接口
export interface AIMessage {
  content: string;
  messageType: MessageType;
  timestamp: number;
  metadata?: Record<string, any>;
}

// AI错误接口
export interface AIError {
  code: string;
  message: string;
  details?: Record<string, any>;
}

// 发送消息请求接口
export interface SendMessageRequest {
  conversationId: string;
  message: string;
  messageType: MessageType;
  aiPersonality: AIPersonality;
  metadata?: Record<string, any>;
}

// 创建对话请求接口
export interface CreateConversationRequest {
  title: string;
  aiPersonality: AIPersonality;
}

// API响应基础接口
export interface BaseResponse<T> {
  success: boolean;
  data: T;
  error?: AIError;
  timestamp: number;
}

// 消息列表响应接口
export interface MessagesResponse extends BaseResponse<MessagesData> {}

export interface MessagesData {
  messages: ChatMessage[];
  hasMore: boolean;
  total: number;
}

// 对话响应接口
export interface ConversationResponse extends BaseResponse<ConversationData> {}

export interface ConversationData {
  conversationId: string;
  title: string;
  createdAt: number;
}

// 对话列表响应接口
export interface ConversationsResponse extends BaseResponse<Conversation[]> {}

// 消息统计接口
export interface MessageStatistics {
  totalMessages: number;
  userMessages: number;
  aiMessages: number;
  averageResponseTime: number;
  lastMessageTime?: number;
}

// 数据库实体类
export class ChatMessageEntity {
  id: string = '';
  conversationId: string = '';
  content: string = '';
  messageType: string = MessageType.TEXT;
  sender: string = MessageSender.USER;
  status: string = MessageStatus.SENDING;
  timestamp: number = 0;
  metadata?: string; // JSON字符串

  constructor(data?: Partial<ChatMessageEntity>) {
    if (data) {
      Object.assign(this, data);
    }
  }

  // 转换为ChatMessage接口
  toChatMessage(): ChatMessage {
    let metadata: Record<string, any> | undefined;
    if (this.metadata) {
      try {
        metadata = JSON.parse(this.metadata);
      } catch (error) {
        console.error('Failed to parse metadata:', error);
      }
    }

    return {
      id: this.id,
      conversationId: this.conversationId,
      content: this.content,
      messageType: this.messageType as MessageType,
      sender: this.sender as MessageSender,
      status: this.status as MessageStatus,
      timestamp: this.timestamp,
      metadata
    };
  }

  // 从ChatMessage创建实体
  static fromChatMessage(message: ChatMessage): ChatMessageEntity {
    const entity = new ChatMessageEntity();
    entity.id = message.id;
    entity.conversationId = message.conversationId;
    entity.content = message.content;
    entity.messageType = message.messageType;
    entity.sender = message.sender;
    entity.status = message.status;
    entity.timestamp = message.timestamp;
    
    if (message.metadata) {
      entity.metadata = JSON.stringify(message.metadata);
    }

    return entity;
  }
}

export class ConversationEntity {
  id: string = '';
  title: string = '';
  aiPersonality: string = AIPersonality.DEFAULT;
  createdAt: number = 0;
  updatedAt: number = 0;
  isArchived: boolean = false;
  messageCount: number = 0;
  lastMessage: string = '';

  constructor(data?: Partial<ConversationEntity>) {
    if (data) {
      Object.assign(this, data);
    }
  }

  // 转换为Conversation接口
  toConversation(): Conversation {
    return {
      id: this.id,
      title: this.title,
      aiPersonality: this.aiPersonality as AIPersonality,
      createdAt: this.createdAt,
      updatedAt: this.updatedAt,
      isArchived: this.isArchived,
      messageCount: this.messageCount,
      lastMessage: this.lastMessage
    };
  }

  // 从Conversation创建实体
  static fromConversation(conversation: Conversation): ConversationEntity {
    const entity = new ConversationEntity();
    entity.id = conversation.id;
    entity.title = conversation.title;
    entity.aiPersonality = conversation.aiPersonality;
    entity.createdAt = conversation.createdAt;
    entity.updatedAt = conversation.updatedAt;
    entity.isArchived = conversation.isArchived;
    entity.messageCount = conversation.messageCount || 0;
    entity.lastMessage = conversation.lastMessage || '';
    return entity;
  }
}

// 工具函数
export class ChatUtils {
  // 生成唯一ID
  static generateId(): string {
    return Date.now().toString(36) + Math.random().toString(36).substr(2);
  }

  // 格式化时间
  static formatTime(timestamp: number): string {
    const date = new Date(timestamp);
    const now = new Date();
    const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    const messageDate = new Date(date.getFullYear(), date.getMonth(), date.getDate());

    if (messageDate.getTime() === today.getTime()) {
      // 今天
      return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
    } else if (messageDate.getTime() === today.getTime() - 24 * 60 * 60 * 1000) {
      // 昨天
      return '昨天 ' + date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
    } else {
      // 其他日期
      return date.toLocaleDateString('zh-CN', { month: '2-digit', day: '2-digit' }) + ' ' +
        date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
    }
  }

  // 获取AI人格显示名称
  static getAIPersonalityDisplayName(personality: AIPersonality): string {
    switch (personality) {
      case AIPersonality.DEFAULT:
        return '默认助手';
      case AIPersonality.WISE_SAGE:
        return '智慧长者';
      case AIPersonality.FRIENDLY_GUIDE:
        return '友善向导';
      case AIPersonality.SCHOLARLY:
        return '学者专家';
      case AIPersonality.POETIC:
        return '诗意文人';
      default:
        return '默认助手';
    }
  }

  // 获取AI人格描述
  static getAIPersonalityDescription(personality: AIPersonality): string {
    switch (personality) {
      case AIPersonality.DEFAULT:
        return '通用AI助手，适合各种日常对话';
      case AIPersonality.WISE_SAGE:
        return '融合古代智慧，以长者口吻提供人生指导';
      case AIPersonality.FRIENDLY_GUIDE:
        return '友善热情，善于解释和指导';
      case AIPersonality.SCHOLARLY:
        return '学术严谨，提供专业深入的分析';
      case AIPersonality.POETIC:
        return '富有诗意，用优美的语言表达思想';
      default:
        return '通用AI助手，适合各种日常对话';
    }
  }

  // 获取消息状态显示文本
  static getMessageStatusText(status: MessageStatus): string {
    switch (status) {
      case MessageStatus.SENDING:
        return '发送中...';
      case MessageStatus.SENT:
        return '已发送';
      case MessageStatus.FAILED:
        return '发送失败';
      case MessageStatus.RECEIVED:
        return '已接收';
      default:
        return '';
    }
  }

  // 验证消息内容
  static validateMessageContent(content: string): { valid: boolean; error?: string } {
    if (!content || content.trim().length === 0) {
      return { valid: false, error: '消息内容不能为空' };
    }

    if (content.length > 2000) {
      return { valid: false, error: '消息内容不能超过2000个字符' };
    }

    return { valid: true };
  }

  // 验证对话标题
  static validateConversationTitle(title: string): { valid: boolean; error?: string } {
    if (!title || title.trim().length === 0) {
      return { valid: false, error: '对话标题不能为空' };
    }

    if (title.length > 100) {
      return { valid: false, error: '对话标题不能超过100个字符' };
    }

    return { valid: true };
  }
}