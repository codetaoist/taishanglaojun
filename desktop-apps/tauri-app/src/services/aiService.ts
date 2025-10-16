import { AIServiceConfig, AIProviderConfig, getAIConfig, saveAIConfig } from '../config/aiConfig';
import { selectAvailableEndpoint } from '../config/apiConfig';
import { isTauriEnvironment } from '../utils/environment';
import '../types/tauri.d';

// AI请求和响应接口
export interface AIRequest {
  id: string;
  provider?: string;
  model?: string;
  capability: string;
  requestType: string;
  input: any;
  context?: any;
  requirements?: string[];
  timeout?: number;
  stream?: boolean;
}

export interface AIResponse {
  requestId: string;
  success: boolean;
  result: any;
  confidence?: number;
  usedCapabilities?: string[];
  metadata?: any;
  error?: string;
  processTime?: number;
  createdAt: string;
  provider: string;
  model: string;
}

export interface ChatMessage {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: Date;
  metadata?: {
    model?: string;
    tokens?: number;
    provider?: string;
    confidence?: number;
    error?: string;
  };
}

export interface ChatSession {
  id: string;
  title: string;
  messages: ChatMessage[];
  createdAt: Date;
  updatedAt: Date;
  model: string;
  provider: string;
}

export interface ImageGenerationRequest {
  prompt: string;
  style?: string;
  size?: string;
  quality?: number;
  steps?: number;
  guidance?: number;
  negativePrompt?: string;
  model?: string;
  provider?: string;
}

export interface ImageAnalysisRequest {
  imageUrl?: string;
  imageData?: string; // base64
  analysisTypes: string[];
  model?: string;
  provider?: string;
}

export interface DocumentProcessingRequest {
  content: string;
  documentType: string;
  tasks: string[];
  model?: string;
  provider?: string;
}

class AIServiceClient {
  private config: AIServiceConfig;
  private responseCache: Map<string, AIResponse> = new Map();
  private metrics: {
    totalRequests: number;
    successfulRequests: number;
    failedRequests: number;
    averageResponseTime: number;
  } = {
    totalRequests: 0,
    successfulRequests: 0,
    failedRequests: 0,
    averageResponseTime: 0
  };

  constructor() {
    this.config = getAIConfig();
  }

  // 更新配置
  updateConfig(newConfig: Partial<AIServiceConfig>): void {
    this.config = { ...this.config, ...newConfig };
    saveAIConfig(this.config);
  }

  // 获取配置
  getConfig(): AIServiceConfig {
    return this.config;
  }

  // 获取可用的提供商
  getAvailableProviders(): AIProviderConfig[] {
    return this.config.providers.filter(p => p.enabled);
  }

  // 获取指定提供商的模型
  getModelsForProvider(providerId: string): any[] {
    const provider = this.config.providers.find(p => p.id === providerId);
    return provider ? provider.models : [];
  }

  // 选择最佳提供商和模型
  private selectProvider(capability: string, preferredProvider?: string): { provider: AIProviderConfig; model: any } | null {
    const availableProviders = this.getAvailableProviders();
    
    // 如果指定了提供商，优先使用
    if (preferredProvider) {
      const provider = availableProviders.find(p => p.id === preferredProvider);
      if (provider) {
        const model = provider.models.find(m => m.capabilities.includes(capability));
        if (model) {
          return { provider, model };
        }
      }
    }

    // 查找支持该能力的提供商和模型
    for (const provider of availableProviders) {
      const model = provider.models.find(m => m.capabilities.includes(capability));
      if (model) {
        return { provider, model };
      }
    }

    return null;
  }

  // 发送AI请求
  async sendRequest(request: AIRequest): Promise<AIResponse> {
    const startTime = Date.now();
    this.metrics.totalRequests++;

    try {
      // 检查缓存
      if (this.config.enableCaching) {
        const cacheKey = this.generateCacheKey(request);
        const cachedResponse = this.responseCache.get(cacheKey);
        if (cachedResponse && Date.now() - new Date(cachedResponse.createdAt).getTime() < this.config.cacheExpiry) {
          return cachedResponse;
        }
      }

      // 选择提供商和模型
      const selection = this.selectProvider(request.capability, request.provider);
      if (!selection) {
        throw new Error(`No available provider for capability: ${request.capability}`);
      }

      const { provider, model } = selection;

      // 检查Tauri环境
      if (isTauriEnvironment()) {
        // 在Tauri环境中，优先使用invoke调用后端
        try {
          const response = await this.invokeBackend(request, provider, model);
          this.metrics.successfulRequests++;
          return response;
        } catch (tauriError) {
          console.warn('Tauri AI service failed, falling back to HTTP API:', tauriError);
          // 如果Tauri调用失败，回退到HTTP API
        }
      }

      // 使用HTTP API或浏览器环境模拟
      const response = await this.callHttpAPI(request, provider, model);
      this.metrics.successfulRequests++;
      return response;
    } catch (error) {
      this.metrics.failedRequests++;
      const errorResponse: AIResponse = {
        requestId: request.id,
        success: false,
        result: null,
        error: error instanceof Error ? error.message : 'Unknown error',
        createdAt: new Date().toISOString(),
        provider: request.provider || 'unknown',
        model: request.model || 'unknown'
      };
      return errorResponse;
    } finally {
      const endTime = Date.now();
      const responseTime = endTime - startTime;
      this.updateAverageResponseTime(responseTime);
    }
  }

  // 调用Tauri后端
  private async invokeBackend(request: AIRequest, provider: AIProviderConfig, model: any): Promise<AIResponse> {
    const { invoke } = await import('@tauri-apps/api/core');
    
    const backendRequest = {
      ...request,
      provider: provider.id,
      model: model.id
    };

    switch (request.capability) {
      case 'chat':
        return await invoke('chat_with_ai', { 
          message: request.input.message,
          chatType: request.requestType,
          sessionId: request.input.sessionId,
          model: model.id,
          provider: provider.id
        });
      
      case 'image-generation':
        return await invoke('generate_image', {
          prompt: request.input.prompt,
          style: request.input.style,
          size: request.input.size,
          model: model.id,
          provider: provider.id
        });
      
      case 'image-analysis':
        return await invoke('analyze_image', {
          imagePath: request.input.imagePath,
          analysisTypes: request.input.analysisTypes,
          model: model.id,
          provider: provider.id
        });
      
      case 'document-processing':
        return await invoke('process_document', {
          content: request.input.content,
          documentType: request.input.documentType,
          tasks: request.input.tasks,
          model: model.id,
          provider: provider.id
        });
      
      default:
        return await invoke('ai_generic_request', { request: backendRequest });
    }
  }

  // HTTP API调用或模拟响应
  private async callHttpAPI(request: AIRequest, provider: AIProviderConfig, model: any): Promise<AIResponse> {
    try {
      // 尝试使用HTTP API
      const endpoint = await selectAvailableEndpoint('ai', 'user');
      
      const response = await fetch(endpoint.url, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('auth_token') || ''}`
        },
        body: JSON.stringify({
          ...request,
          provider: provider.id,
          model: model.id
        })
      });

      if (response.ok) {
        const apiResponse = await response.json();
        if (apiResponse.success) {
          return apiResponse.data;
        }
      }
      
      // 如果HTTP API失败，回退到模拟响应
      console.warn('HTTP AI API failed, using simulation');
    } catch (error) {
      console.warn('HTTP AI API error, using simulation:', error);
    }

    // 模拟响应（浏览器环境或API不可用时）
    return this.simulateResponse(request, provider, model);
  }

  // 模拟响应（浏览器环境）
  private async simulateResponse(request: AIRequest, provider: AIProviderConfig, model: any): Promise<AIResponse> {
    // 模拟网络延迟
    await new Promise(resolve => setTimeout(resolve, Math.random() * 1000 + 500));

    const baseResponse: AIResponse = {
      requestId: request.id,
      success: true,
      result: null,
      confidence: 0.85 + Math.random() * 0.15,
      usedCapabilities: [request.capability],
      metadata: {
        model: model.id,
        provider: provider.id,
        simulated: true
      },
      createdAt: new Date().toISOString(),
      provider: provider.id,
      model: model.id,
      processTime: Math.floor(Math.random() * 2000 + 500)
    };

    switch (request.capability) {
      case 'chat':
        baseResponse.result = {
          content: `这是来自${model.name}的模拟回复：${request.input.message}`,
          messageId: `msg_${Date.now()}`,
          sessionId: request.input.sessionId || `session_${Date.now()}`,
          tokensUsed: Math.floor(Math.random() * 100 + 50)
        };
        break;

      case 'image-generation':
        baseResponse.result = {
          imageUrl: `data:image/svg+xml;base64,${btoa(`
            <svg width="512" height="512" xmlns="http://www.w3.org/2000/svg">
              <rect width="100%" height="100%" fill="#f0f0f0"/>
              <text x="50%" y="50%" text-anchor="middle" dy=".3em" font-family="Arial" font-size="24">
                模拟图像: ${request.input.prompt}
              </text>
            </svg>
          `)}`,
          prompt: request.input.prompt,
          style: request.input.style || 'default',
          size: request.input.size || '512x512',
          seed: Math.floor(Math.random() * 1000000)
        };
        break;

      case 'image-analysis':
        baseResponse.result = {
          description: '这是一个模拟的图像分析结果',
          objects: ['object1', 'object2'],
          colors: ['#ff0000', '#00ff00', '#0000ff'],
          tags: ['tag1', 'tag2', 'tag3'],
          confidence: 0.9
        };
        break;

      case 'document-processing':
        baseResponse.result = {
          summary: '这是文档的模拟摘要',
          keyPoints: ['要点1', '要点2', '要点3'],
          sentiment: 'positive',
          entities: ['实体1', '实体2'],
          processedTasks: request.input.tasks
        };
        break;

      default:
        baseResponse.result = {
          message: `模拟的${request.capability}响应`,
          data: request.input
        };
    }

    // 缓存响应
    if (this.config.enableCaching) {
      const cacheKey = this.generateCacheKey(request);
      this.responseCache.set(cacheKey, baseResponse);
    }

    return baseResponse;
  }

  // 聊天相关方法
  async sendChatMessage(message: string, sessionId?: string, options?: {
    model?: string;
    provider?: string;
    temperature?: number;
    maxTokens?: number;
  }): Promise<AIResponse> {
    const request: AIRequest = {
      id: `chat_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      capability: 'chat',
      requestType: 'message',
      input: {
        message,
        sessionId,
        ...options
      },
      provider: options?.provider,
      model: options?.model
    };

    return this.sendRequest(request);
  }

  // 图像生成方法
  async generateImage(request: ImageGenerationRequest): Promise<AIResponse> {
    const aiRequest: AIRequest = {
      id: `img_gen_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      capability: 'image-generation',
      requestType: 'generate',
      input: request,
      provider: request.provider,
      model: request.model
    };

    return this.sendRequest(aiRequest);
  }

  // 图像分析方法
  async analyzeImage(request: ImageAnalysisRequest): Promise<AIResponse> {
    const aiRequest: AIRequest = {
      id: `img_analysis_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      capability: 'image-analysis',
      requestType: 'analyze',
      input: request,
      provider: request.provider,
      model: request.model
    };

    return this.sendRequest(aiRequest);
  }

  // 文档处理方法
  async processDocument(request: DocumentProcessingRequest): Promise<AIResponse> {
    const aiRequest: AIRequest = {
      id: `doc_proc_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      capability: 'document-processing',
      requestType: 'process',
      input: request,
      provider: request.provider,
      model: request.model
    };

    return this.sendRequest(aiRequest);
  }

  // 获取系统状态
  async getSystemStatus(): Promise<any> {
    if (typeof window !== 'undefined' && window.__TAURI__) {
      const { invoke } = await import('@tauri-apps/api/core');
      return await invoke('get_system_status');
    } else {
      // 模拟系统状态
      return {
        cpu_usage: Math.random() * 0.8,
        memory_usage: Math.random() * 0.7,
        disk_usage: Math.random() * 0.6,
        network_status: 'connected',
        ai_service_status: 'running',
        database_status: 'connected'
      };
    }
  }

  // 获取性能指标
  getMetrics() {
    return {
      ...this.metrics,
      successRate: this.metrics.totalRequests > 0 
        ? this.metrics.successfulRequests / this.metrics.totalRequests 
        : 0
    };
  }

  // 健康检查
  async healthCheck(): Promise<boolean> {
    try {
      const status = await this.getSystemStatus();
      return status.ai_service_status === 'running';
    } catch (error) {
      return false;
    }
  }

  // 工具方法
  private generateCacheKey(request: AIRequest): string {
    return `${request.capability}_${request.requestType}_${JSON.stringify(request.input)}`;
  }

  private updateAverageResponseTime(responseTime: number): void {
    if (this.metrics.totalRequests === 1) {
      this.metrics.averageResponseTime = responseTime;
    } else {
      this.metrics.averageResponseTime = 
        (this.metrics.averageResponseTime * (this.metrics.totalRequests - 1) + responseTime) / this.metrics.totalRequests;
    }
  }

  // 清理缓存
  clearCache(): void {
    this.responseCache.clear();
  }

  // 重置指标
  resetMetrics(): void {
    this.metrics = {
      totalRequests: 0,
      successfulRequests: 0,
      failedRequests: 0,
      averageResponseTime: 0
    };
  }

  // 输入优化相关方法
  async optimizeInput(
    text: string, 
    targetAudience?: string, 
    optimizationType?: string, 
    language?: string
  ): Promise<any> {
    if (typeof window !== 'undefined' && window.__TAURI__) {
      const { invoke } = await import('@tauri-apps/api/core');
      return await invoke('optimize_input', {
        text,
        targetAudience,
        optimizationType,
        language
      });
    } else {
      // 模拟优化结果
      return {
        original_text: text,
        best_suggestion: {
          optimized_text: `优化后的文本: ${text}`,
          confidence: 0.85,
          optimization_type: optimizationType || 'clarity',
          changes: ['语法修正', '表达优化']
        },
        suggestions: [
          {
            optimized_text: `建议1: ${text}`,
            confidence: 0.80,
            optimization_type: 'clarity',
            changes: ['语法修正']
          }
        ],
        intent: {
          type: 'question',
          confidence: 0.75,
          context: 'general'
        }
      };
    }
  }

  async optimizeInputForPlatform(text: string, platform: 'windows' | 'macos' | 'linux'): Promise<string> {
    if (typeof window !== 'undefined' && window.__TAURI__) {
      const { invoke } = await import('@tauri-apps/api/core');
      const commandMap = {
        windows: 'optimize_input_windows',
        macos: 'optimize_input_macos',
        linux: 'optimize_input_linux'
      };
      return await invoke(commandMap[platform], { text });
    } else {
      // 模拟平台特定优化
      const platformOptimizations = {
        windows: `[Windows优化] ${text}`,
        macos: `[macOS优化] ${text}`,
        linux: `[Linux优化] ${text}`
      };
      return platformOptimizations[platform];
    }
  }

  async getQuickSuggestions(text: string): Promise<string[]> {
    if (typeof window !== 'undefined' && window.__TAURI__) {
      const { invoke } = await import('@tauri-apps/api/core');
      return await invoke('get_quick_suggestions', { text });
    } else {
      // 模拟快速建议
      return [
        `${text} - 建议1`,
        `${text} - 建议2`,
        `${text} - 建议3`
      ];
    }
  }

  async detectInputIntent(text: string): Promise<any> {
    if (typeof window !== 'undefined' && window.__TAURI__) {
      const { invoke } = await import('@tauri-apps/api/core');
      return await invoke('detect_input_intent', { text });
    } else {
      // 模拟意图检测
      return {
        type: 'question',
        confidence: 0.75,
        context: 'general',
        suggested_actions: ['search', 'clarify']
      };
    }
  }
}

// 导出单例实例
export const aiService = new AIServiceClient();
export default aiService;