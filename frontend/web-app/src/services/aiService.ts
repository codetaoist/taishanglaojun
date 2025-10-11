import axios from 'axios';
import type { AxiosInstance, AxiosResponse } from 'axios';

// AI服务响应接口
export interface ApiResponse<T> {
  code: string;
  message: string;
  data: T;
}

// 图像生成请求接口
export interface ImageGenerationRequest {
  prompt: string;
  style?: string;
  size?: string;
  quality?: number;
  steps?: number;
  guidance?: number;
  negativePrompt?: string;
}

// 图像生成响应接口
export interface ImageGenerationResponse {
  id: string;
  url: string;
  prompt: string;
  style: string;
  size: string;
  timestamp: string;
  metadata: {
    steps: number;
    guidance: number;
    quality: number;
    seed: number;
  };
}

// 图像分析请求接口
export interface ImageAnalysisRequest {
  imageUrl?: string;
  imageFile?: File;
  analysisTypes: string[];
}

// 图像分析响应接口
export interface ImageAnalysisResponse {
  id: string;
  imageUrl: string;
  fileName: string;
  fileSize: string;
  timestamp: string;
  results: {
    objects?: Array<{ name: string; confidence: number; bbox: number[] }>;
    faces?: Array<{ age: number; gender: string; emotion: string; confidence: number }>;
    text?: Array<{ text: string; confidence: number; bbox: number[] }>;
    colors?: Array<{ color: string; percentage: number; hex: string }>;
    tags?: Array<{ tag: string; confidence: number }>;
    description?: string;
    similarity?: Array<{ imageUrl: string; similarity: number; description: string }>;
  };
}

// AI聊天消息接口
export interface ChatMessage {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  timestamp: Date;
  metadata?: {
    model?: string;
    tokens?: number;
    provider?: string;
  };
}

// AI聊天会话接口
export interface ChatSession {
  id: string;
  title: string;
  messages: ChatMessage[];
  createdAt: Date;
  updatedAt: Date;
  model: string;
  provider: string;
}

class AIService {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: import.meta.env.VITE_AI_API_BASE_URL || '/api/ai',
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // 请求拦截器
    this.client.interceptors.request.use(
      (config) => {
        const token = localStorage.getItem('token');
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => {
        return Promise.reject(error);
      }
    );

    // 响应拦截器
    this.client.interceptors.response.use(
      (response: AxiosResponse) => {
        return response;
      },
      (error) => {
        console.error('AI API Error:', error);
        return Promise.reject(error);
      }
    );
  }

  // 图像生成相关API
  async generateImage(request: ImageGenerationRequest): Promise<ApiResponse<ImageGenerationResponse>> {
    const response = await this.client.post('/image/generate', request);
    return response.data;
  }

  async getGenerationHistory(limit?: number): Promise<ApiResponse<ImageGenerationResponse[]>> {
    const response = await this.client.get('/image/history', { 
      params: { limit: limit || 20 } 
    });
    return response.data;
  }

  async deleteGeneratedImage(imageId: string): Promise<ApiResponse<void>> {
    const response = await this.client.delete(`/image/generated/${imageId}`);
    return response.data;
  }

  // 图像分析相关API
  async analyzeImage(request: ImageAnalysisRequest): Promise<ApiResponse<ImageAnalysisResponse>> {
    const formData = new FormData();
    
    if (request.imageFile) {
      formData.append('image', request.imageFile);
    } else if (request.imageUrl) {
      formData.append('imageUrl', request.imageUrl);
    }
    
    formData.append('analysisTypes', JSON.stringify(request.analysisTypes));

    const response = await this.client.post('/image/analyze', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  }

  async getAnalysisHistory(limit?: number): Promise<ApiResponse<ImageAnalysisResponse[]>> {
    const response = await this.client.get('/image/analysis-history', { 
      params: { limit: limit || 20 } 
    });
    return response.data;
  }

  async deleteAnalysisResult(analysisId: string): Promise<ApiResponse<void>> {
    const response = await this.client.delete(`/image/analysis/${analysisId}`);
    return response.data;
  }

  // AI聊天相关API
  async sendChatMessage(
    message: string, 
    sessionId?: string, 
    options?: {
      model?: string;
      provider?: string;
      temperature?: number;
      maxTokens?: number;
    }
  ): Promise<ApiResponse<{
    sessionId: string;
    messageId: string;
    content: string;
    tokensUsed: number;
    model: string;
    provider: string;
  }>> {
    const response = await this.client.post('/chat/message', {
      message,
      sessionId,
      ...options,
    });
    return response.data;
  }

  async getChatSessions(): Promise<ApiResponse<ChatSession[]>> {
    const response = await this.client.get('/chat/sessions');
    return response.data;
  }

  async getChatSession(sessionId: string): Promise<ApiResponse<ChatSession>> {
    const response = await this.client.get(`/chat/sessions/${sessionId}`);
    return response.data;
  }

  async createChatSession(title?: string): Promise<ApiResponse<ChatSession>> {
    const response = await this.client.post('/chat/sessions', { title });
    return response.data;
  }

  async deleteChatSession(sessionId: string): Promise<ApiResponse<void>> {
    const response = await this.client.delete(`/chat/sessions/${sessionId}`);
    return response.data;
  }

  // AI模型和提供商相关API
  async getAvailableProviders(): Promise<ApiResponse<{
    providers: Array<{
      id: string;
      name: string;
      description: string;
      models: Array<{
        id: string;
        name: string;
        description: string;
        capabilities: string[];
        pricing?: {
          inputTokens: number;
          outputTokens: number;
        };
      }>;
    }>;
  }>> {
    const response = await this.client.get('/providers');
    return response.data;
  }

  async getAIHealth(): Promise<ApiResponse<{
    status: 'healthy' | 'degraded' | 'down';
    providers: Record<string, {
      status: 'online' | 'offline';
      latency?: number;
      lastCheck: string;
    }>;
    uptime: number;
  }>> {
    const response = await this.client.get('/health');
    return response.data;
  }

  // 用户配额和统计相关API
  async getUserQuota(): Promise<ApiResponse<{
    daily: {
      used: number;
      limit: number;
      remaining: number;
    };
    monthly: {
      used: number;
      limit: number;
      remaining: number;
    };
    resetTime: string;
  }>> {
    const response = await this.client.get('/quota');
    return response.data;
  }

  async getUserStats(): Promise<ApiResponse<{
    totalGenerations: number;
    totalAnalyses: number;
    totalChatMessages: number;
    favoriteModels: Array<{
      model: string;
      provider: string;
      usage: number;
    }>;
    recentActivity: Array<{
      type: 'generation' | 'analysis' | 'chat';
      timestamp: string;
      details: any;
    }>;
  }>> {
    const response = await this.client.get('/stats');
    return response.data;
  }
}

// 创建AI服务实例
export const aiService = new AIService();

// 导出类型
export type {
  ImageGenerationRequest,
  ImageGenerationResponse,
  ImageAnalysisRequest,
  ImageAnalysisResponse,
  ChatMessage,
  ChatSession,
};

export default aiService;