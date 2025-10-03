/**
 * AI服务接口和实现
 * 负责与AI后端服务的通信
 */

import http from '@ohos.net.http';
import { 
  AIResponse, 
  SendMessageRequest, 
  CreateConversationRequest,
  MessagesResponse,
  ConversationResponse,
  ConversationsResponse,
  MessagesData,
  ConversationData,
  Conversation,
  AIPersonality,
  MessageType,
  AIMessage,
  AIError
} from '../model/ChatMessage';

// AI服务接口
export interface AIServiceInterface {
  sendMessage(request: SendMessageRequest): Promise<AIResponse>;
  getMessages(conversationId: string, page?: number, limit?: number, before?: number): Promise<MessagesData>;
  createConversation(request: CreateConversationRequest): Promise<ConversationData>;
  getConversations(): Promise<Conversation[]>;
  deleteConversation(conversationId: string): Promise<void>;
  sendMessageStream(request: SendMessageRequest): AsyncGenerator<string, void, unknown>;
}

// AI服务实现
export class AIService implements AIServiceInterface {
  private static instance: AIService;
  private baseURL: string = 'https://api.taishanglaojun.com';
  private timeout: number = 30000;
  private maxRetries: number = 3;

  private constructor() {}

  static getInstance(): AIService {
    if (!AIService.instance) {
      AIService.instance = new AIService();
    }
    return AIService.instance;
  }

  // 发送消息
  async sendMessage(request: SendMessageRequest): Promise<AIResponse> {
    const url = `${this.baseURL}/api/ai/chat/send`;
    
    try {
      const response = await this.makeRequest('POST', url, request);
      return this.parseResponse<AIResponse>(response);
    } catch (error) {
      console.error('Send message failed:', error);
      throw this.handleError(error);
    }
  }

  // 获取消息列表
  async getMessages(
    conversationId: string, 
    page: number = 1, 
    limit: number = 50, 
    before?: number
  ): Promise<MessagesData> {
    let url = `${this.baseURL}/api/ai/chat/conversations/${conversationId}/messages?page=${page}&limit=${limit}`;
    
    if (before) {
      url += `&before=${before}`;
    }

    try {
      const response = await this.makeRequest('GET', url);
      const result = this.parseResponse<MessagesResponse>(response);
      
      if (result.success) {
        return result.data;
      } else {
        throw new Error(result.error?.message || 'Failed to get messages');
      }
    } catch (error) {
      console.error('Get messages failed:', error);
      throw this.handleError(error);
    }
  }

  // 创建对话
  async createConversation(request: CreateConversationRequest): Promise<ConversationData> {
    const url = `${this.baseURL}/api/ai/chat/conversations`;
    
    try {
      const response = await this.makeRequest('POST', url, request);
      const result = this.parseResponse<ConversationResponse>(response);
      
      if (result.success) {
        return result.data;
      } else {
        throw new Error(result.error?.message || 'Failed to create conversation');
      }
    } catch (error) {
      console.error('Create conversation failed:', error);
      throw this.handleError(error);
    }
  }

  // 获取对话列表
  async getConversations(): Promise<Conversation[]> {
    const url = `${this.baseURL}/api/ai/chat/conversations`;
    
    try {
      const response = await this.makeRequest('GET', url);
      const result = this.parseResponse<ConversationsResponse>(response);
      
      if (result.success) {
        return result.data;
      } else {
        throw new Error(result.error?.message || 'Failed to get conversations');
      }
    } catch (error) {
      console.error('Get conversations failed:', error);
      throw this.handleError(error);
    }
  }

  // 删除对话
  async deleteConversation(conversationId: string): Promise<void> {
    const url = `${this.baseURL}/api/ai/chat/conversations/${conversationId}`;
    
    try {
      await this.makeRequest('DELETE', url);
    } catch (error) {
      console.error('Delete conversation failed:', error);
      throw this.handleError(error);
    }
  }

  // 流式发送消息（模拟实现）
  async* sendMessageStream(request: SendMessageRequest): AsyncGenerator<string, void, unknown> {
    try {
      const response = await this.sendMessage(request);
      
      if (response.message?.content) {
        const words = response.message.content.split(' ');
        
        for (const word of words) {
          yield word + ' ';
          // 模拟延迟
          await new Promise(resolve => setTimeout(resolve, 100));
        }
      }
    } catch (error) {
      console.error('Stream message failed:', error);
      throw this.handleError(error);
    }
  }

  // 发起HTTP请求
  private async makeRequest(
    method: 'GET' | 'POST' | 'PUT' | 'DELETE',
    url: string,
    data?: any
  ): Promise<http.HttpResponse> {
    return new Promise((resolve, reject) => {
      const httpRequest = http.createHttp();
      
      const options: http.HttpRequestOptions = {
        method: method,
        header: {
          'Content-Type': 'application/json',
          'Authorization': this.getAuthToken() ? `Bearer ${this.getAuthToken()}` : ''
        },
        readTimeout: this.timeout,
        connectTimeout: this.timeout
      };

      if (data && (method === 'POST' || method === 'PUT')) {
        options.extraData = JSON.stringify(data);
      }

      httpRequest.request(url, options, (err, data) => {
        if (err) {
          console.error('HTTP request failed:', err);
          reject(err);
        } else {
          resolve(data);
        }
        
        httpRequest.destroy();
      });
    });
  }

  // 解析响应
  private parseResponse<T>(response: http.HttpResponse): T {
    try {
      if (typeof response.result === 'string') {
        return JSON.parse(response.result) as T;
      } else {
        return response.result as T;
      }
    } catch (error) {
      console.error('Failed to parse response:', error);
      throw new Error('Invalid response format');
    }
  }

  // 获取认证令牌
  private getAuthToken(): string | null {
    // 从本地存储获取认证令牌
    // 这里需要根据实际的存储方式实现
    return null;
  }

  // 错误处理
  private handleError(error: any): Error {
    if (error instanceof Error) {
      return error;
    }
    
    if (typeof error === 'string') {
      return new Error(error);
    }
    
    return new Error('Unknown error occurred');
  }
}

// Mock AI服务（用于测试）
export class MockAIService implements AIServiceInterface {
  private delay: number;

  constructor(delay: number = 1000) {
    this.delay = delay;
  }

  async sendMessage(request: SendMessageRequest): Promise<AIResponse> {
    await this.sleep(this.delay);
    
    const mockResponse: AIResponse = {
      success: true,
      message: {
        content: this.generateMockResponse(request.message, request.aiPersonality),
        messageType: MessageType.TEXT,
        timestamp: Date.now(),
        metadata: {
          model: 'taishang-v1',
          tokens_used: Math.floor(Math.random() * 100) + 20,
          response_time: Math.floor(Math.random() * 1200) + 800
        }
      },
      suggestions: this.generateSuggestions(request.aiPersonality)
    };

    return mockResponse;
  }

  async getMessages(
    conversationId: string, 
    page: number = 1, 
    limit: number = 50, 
    before?: number
  ): Promise<MessagesData> {
    await this.sleep(this.delay);
    
    return {
      messages: [],
      hasMore: false,
      total: 0
    };
  }

  async createConversation(request: CreateConversationRequest): Promise<ConversationData> {
    await this.sleep(this.delay);
    
    return {
      conversationId: this.generateId(),
      title: request.title,
      createdAt: Date.now()
    };
  }

  async getConversations(): Promise<Conversation[]> {
    await this.sleep(this.delay);
    return [];
  }

  async deleteConversation(conversationId: string): Promise<void> {
    await this.sleep(this.delay);
  }

  async* sendMessageStream(request: SendMessageRequest): AsyncGenerator<string, void, unknown> {
    const response = this.generateMockResponse(request.message, request.aiPersonality);
    const words = response.split(' ');
    
    for (const word of words) {
      yield word + ' ';
      await this.sleep(100);
    }
  }

  private generateMockResponse(message: string, personality: AIPersonality): string {
    switch (personality) {
      case AIPersonality.DEFAULT:
        return `感谢您的提问。关于「${message}」，我认为这是一个很有意思的话题。`;
      case AIPersonality.WISE_SAGE:
        return `古人云：「${message}」，此言深有道理。老夫以为，万事皆有其道，需细细参悟。`;
      case AIPersonality.FRIENDLY_GUIDE:
        return `您好！关于「${message}」这个问题，让我来为您详细解答一下。`;
      case AIPersonality.SCHOLARLY:
        return `根据相关研究和理论分析，「${message}」这一概念具有重要的学术价值和实践意义。`;
      case AIPersonality.POETIC:
        return `「${message}」如春风化雨，润物无声。正所谓：山重水复疑无路，柳暗花明又一村。`;
      default:
        return `关于您的问题「${message}」，我来为您解答。`;
    }
  }

  private generateSuggestions(personality: AIPersonality): string[] {
    switch (personality) {
      case AIPersonality.DEFAULT:
        return ['请问您还有其他问题吗？', '我可以为您详细解释', '需要更多信息吗？'];
      case AIPersonality.WISE_SAGE:
        return ['可否请教更深层的道理？', '古籍中还有何见解？', '此理可有实践之法？'];
      case AIPersonality.FRIENDLY_GUIDE:
        return ['还有什么我可以帮助您的？', '您想了解更多细节吗？', '有其他相关问题吗？'];
      case AIPersonality.SCHOLARLY:
        return ['是否需要相关文献资料？', '可以进一步分析吗？', '有实证研究支持吗？'];
      case AIPersonality.POETIC:
        return ['可否再吟一首？', '此情此景，还有何感？', '诗意人生，何处寻觅？'];
      default:
        return ['还有什么问题吗？'];
    }
  }

  private generateId(): string {
    return Date.now().toString(36) + Math.random().toString(36).substr(2);
  }

  private sleep(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}

// AI服务错误类
export class AIServiceError extends Error {
  code: string;
  details?: Record<string, any>;

  constructor(code: string, message: string, details?: Record<string, any>) {
    super(message);
    this.name = 'AIServiceError';
    this.code = code;
    this.details = details;
  }

  static invalidURL(): AIServiceError {
    return new AIServiceError('INVALID_URL', '无效的URL');
  }

  static networkError(): AIServiceError {
    return new AIServiceError('NETWORK_ERROR', '网络连接错误');
  }

  static serverError(message: string): AIServiceError {
    return new AIServiceError('SERVER_ERROR', `服务器错误: ${message}`);
  }

  static noContent(): AIServiceError {
    return new AIServiceError('NO_CONTENT', '没有内容');
  }

  static decodingError(): AIServiceError {
    return new AIServiceError('DECODING_ERROR', '数据解析错误');
  }
}