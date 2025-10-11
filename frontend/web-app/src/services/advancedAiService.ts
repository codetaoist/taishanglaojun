import axios from 'axios';
import type { AxiosInstance, AxiosResponse } from 'axios';

// 基础响应类型
interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
    details?: string;
  };
  message?: string;
  timestamp: string;
  request_id: string;
}

// AGI 相关类型
interface ReasoningRequest {
  query: string;
  context?: string;
  reasoning_type?: 'deductive' | 'inductive' | 'abductive';
  max_steps?: number;
  confidence_threshold?: number;
}

interface ReasoningResponse {
  solution: string;
  reasoning_steps: Array<{
    step: number;
    description: string;
    confidence: number;
  }>;
  confidence: number;
  alternatives?: string[];
}

interface PlanningRequest {
  goal: string;
  constraints?: Record<string, any>;
  context?: Record<string, any>;
  requirements?: {
    detail_level?: 'low' | 'medium' | 'high';
    include_risks?: boolean;
    include_timeline?: boolean;
  };
  priority?: number;
  timeout?: number;
}

interface PlanningResponse {
  plan: {
    id: string;
    goal: string;
    steps: Array<{
      id: string;
      title: string;
      description: string;
      dependencies: string[];
      estimated_time: string;
      resources: string[];
      risks: string[];
    }>;
    timeline: string;
    total_cost: number;
    success_probability: number;
  };
  alternatives?: any[];
}

interface MultimodalRequest {
  inputs: Array<{
    type: 'text' | 'image' | 'audio' | 'video';
    content: string; // base64 for media files
    metadata?: Record<string, any>;
  }>;
  task: string;
  fusion_strategy?: 'early' | 'late' | 'hybrid';
  output_format?: 'text' | 'structured' | 'multimodal';
}

interface MultimodalResponse {
  result: {
    content: string;
    confidence: number;
    modality_contributions: Record<string, number>;
  };
  analysis: {
    detected_objects?: string[];
    sentiment?: string;
    topics?: string[];
    relationships?: Array<{
      source: string;
      target: string;
      type: string;
      confidence: number;
    }>;
  };
}

// 元学习相关类型
interface MetaLearningRequest {
  task_type: string;
  domain: string;
  data: Array<{
    input: any;
    label?: any;
    metadata?: Record<string, any>;
  }>;
  strategy?: 'few_shot' | 'model_agnostic' | 'gradient_based';
  parameters?: Record<string, any>;
}

interface MetaLearningResponse {
  strategy: {
    id: string;
    algorithm: string;
    performance_metrics: {
      accuracy: number;
      adaptation_speed: number;
      generalization: number;
    };
    learned_parameters: Record<string, any>;
    training_time: number;
  };
}

interface AdaptationRequest {
  base_strategy_id: string;
  new_task: {
    description: string;
    sample_data: any;
    target_metric: string;
  };
  adaptation_budget?: {
    max_iterations?: number;
    max_time?: string;
  };
}

// 自我进化相关类型
interface PerformanceMetrics {
  accuracy: number;
  latency: number;
  throughput: number;
  resource_usage: {
    cpu: number;
    memory: number;
    gpu?: number;
  };
  user_satisfaction: number;
  error_rate: number;
}

interface OptimizationRequest {
  targets: string[];
  strategy?: 'genetic' | 'gradient' | 'reinforcement' | 'hybrid';
  constraints?: Record<string, any>;
  budget?: {
    max_iterations?: number;
    max_time?: string;
    max_resources?: Record<string, number>;
  };
}

interface OptimizationResponse {
  optimization_id: string;
  status: 'running' | 'completed' | 'failed';
  progress: number;
  current_metrics: PerformanceMetrics;
  improvements: Record<string, number>;
  estimated_completion: string;
}

// WebSocket 消息类型
interface WebSocketMessage {
  type: 'reasoning' | 'learning' | 'evolution' | 'status';
  data: any;
  timestamp: string;
  request_id?: string;
}

class AdvancedAIService {
  private client: AxiosInstance;
  private wsConnections: Map<string, WebSocket> = new Map();

  constructor() {
    this.client = axios.create({
      baseURL: import.meta.env.VITE_ADVANCED_AI_API_URL || 'http://localhost:8080/api/v1/advanced-ai',
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
      (error) => Promise.reject(error)
    );

    // 响应拦截器
    this.client.interceptors.response.use(
      (response: AxiosResponse<ApiResponse>) => response,
      (error) => {
        console.error('Advanced AI API Error:', error);
        return Promise.reject(error);
      }
    );
  }

  // AGI 能力接口
  async reasoning(request: ReasoningRequest): Promise<ApiResponse<ReasoningResponse>> {
    const response = await this.client.post('/agi/reasoning', request);
    return response.data;
  }

  async planning(request: PlanningRequest): Promise<ApiResponse<PlanningResponse>> {
    const response = await this.client.post('/agi/planning', request);
    return response.data;
  }

  async multimodal(request: MultimodalRequest): Promise<ApiResponse<MultimodalResponse>> {
    const response = await this.client.post('/agi/multimodal', request);
    return response.data;
  }

  async creativity(prompt: string, options?: {
    type?: 'text' | 'image' | 'code' | 'music';
    style?: string;
    constraints?: Record<string, any>;
  }): Promise<ApiResponse<any>> {
    const response = await this.client.post('/agi/creativity', {
      prompt,
      ...options,
    });
    return response.data;
  }

  // 元学习接口
  async metaLearning(request: MetaLearningRequest): Promise<ApiResponse<MetaLearningResponse>> {
    const response = await this.client.post('/meta-learning/learn', request);
    return response.data;
  }

  async adaptation(request: AdaptationRequest): Promise<ApiResponse<any>> {
    const response = await this.client.post('/meta-learning/adapt', request);
    return response.data;
  }

  async knowledgeTransfer(sourceTaskId: string, targetTaskId: string, options?: {
    transfer_method?: 'feature' | 'parameter' | 'gradient';
    similarity_threshold?: number;
  }): Promise<ApiResponse<any>> {
    const response = await this.client.post('/meta-learning/transfer', {
      source_task_id: sourceTaskId,
      target_task_id: targetTaskId,
      ...options,
    });
    return response.data;
  }

  // 自我进化接口
  async getPerformanceMetrics(): Promise<ApiResponse<PerformanceMetrics>> {
    const response = await this.client.get('/self-evolution/performance');
    return response.data;
  }

  async triggerOptimization(request: OptimizationRequest): Promise<ApiResponse<OptimizationResponse>> {
    const response = await this.client.post('/self-evolution/optimize', request);
    return response.data;
  }

  async getOptimizationStatus(optimizationId: string): Promise<ApiResponse<OptimizationResponse>> {
    const response = await this.client.get(`/self-evolution/optimize/${optimizationId}`);
    return response.data;
  }

  async getEvolutionHistory(limit?: number): Promise<ApiResponse<any[]>> {
    const response = await this.client.get('/self-evolution/history', {
      params: { limit },
    });
    return response.data;
  }

  // 系统管理接口
  async getSystemStatus(): Promise<ApiResponse<any>> {
    const response = await this.client.get('/system/status');
    return response.data;
  }

  async getSystemMetrics(): Promise<ApiResponse<any>> {
    const response = await this.client.get('/system/metrics');
    return response.data;
  }

  async updateSystemConfig(config: Record<string, any>): Promise<ApiResponse<any>> {
    const response = await this.client.put('/system/config', config);
    return response.data;
  }

  // WebSocket 连接管理
  connectWebSocket(endpoint: string, onMessage: (message: WebSocketMessage) => void): string {
    const wsUrl = `${import.meta.env.VITE_ADVANCED_AI_WS_URL || 'ws://localhost:8080'}/ws/${endpoint}`;
    const connectionId = `${endpoint}_${Date.now()}`;
    
    const ws = new WebSocket(wsUrl);
    
    ws.onopen = () => {
      console.log(`WebSocket connected: ${endpoint}`);
    };
    
    ws.onmessage = (event) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data);
        onMessage(message);
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error);
      }
    };
    
    ws.onclose = () => {
      console.log(`WebSocket disconnected: ${endpoint}`);
      this.wsConnections.delete(connectionId);
    };
    
    ws.onerror = (error) => {
      console.error(`WebSocket error: ${endpoint}`, error);
    };
    
    this.wsConnections.set(connectionId, ws);
    return connectionId;
  }

  disconnectWebSocket(connectionId: string): void {
    const ws = this.wsConnections.get(connectionId);
    if (ws) {
      ws.close();
      this.wsConnections.delete(connectionId);
    }
  }

  // 批量处理接口
  async batchProcess(requests: Array<{
    type: 'reasoning' | 'planning' | 'multimodal' | 'creativity';
    data: any;
    priority?: number;
  }>): Promise<ApiResponse<any[]>> {
    const response = await this.client.post('/batch/process', { requests });
    return response.data;
  }

  // 模型管理接口
  async listAvailableModels(): Promise<ApiResponse<any[]>> {
    const response = await this.client.get('/models');
    return response.data;
  }

  async getModelInfo(modelId: string): Promise<ApiResponse<any>> {
    const response = await this.client.get(`/models/${modelId}`);
    return response.data;
  }

  async switchModel(modelId: string, capability: string): Promise<ApiResponse<any>> {
    const response = await this.client.post(`/models/${modelId}/switch`, { capability });
    return response.data;
  }
}

// 导出单例实例
export const advancedAIService = new AdvancedAIService();
export default advancedAIService;

// 导出类型
export type {
  ApiResponse,
  ReasoningRequest,
  ReasoningResponse,
  PlanningRequest,
  PlanningResponse,
  MultimodalRequest,
  MultimodalResponse,
  MetaLearningRequest,
  MetaLearningResponse,
  AdaptationRequest,
  PerformanceMetrics,
  OptimizationRequest,
  OptimizationResponse,
  WebSocketMessage,
};