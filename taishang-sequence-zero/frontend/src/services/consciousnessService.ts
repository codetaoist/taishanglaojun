import { api } from './api';

// 意识状态接口
export interface ConsciousnessState {
  id: number;
  user_id: number;
  emotional_state: Record<string, number>;
  cognitive_level: number;
  spiritual_depth: number;
  personality_traits: Record<string, number>;
  consciousness_type: string;
  fusion_level: number;
  last_updated: string;
  metadata: Record<string, any>;
}

// 意识融合请求接口
export interface FusionRequest {
  user_id: number;
  target_state: string;
  intensity: number;
  duration?: number;
  personalization?: Record<string, any>;
}

// 意识分析请求接口
export interface AnalysisRequest {
  user_id: number;
  input_data: Record<string, any>;
  analysis_type: string;
}

// 个性化适配请求接口
export interface AdaptationRequest {
  user_id: number;
  preferences: Record<string, any>;
  goals?: string[];
}

// 融合历史记录接口
export interface FusionHistory {
  id: number;
  user_id: number;
  session_type: string;
  start_time: string;
  end_time?: string;
  effectiveness: number;
  feedback: string;
  metrics: Record<string, any>;
}

// 意识融合相关接口
export const consciousnessService = {
  // 分析意识状态
  analyzeConsciousness: async (request: AnalysisRequest) => {
    const response = await api.post('/consciousness/analyze', request);
    return response.data;
  },

  // 执行意识融合
  fuseConsciousness: async (request: FusionRequest) => {
    const response = await api.post('/consciousness/fuse', request);
    return response.data;
  },

  // 获取意识状态
  getConsciousnessState: async (userId: number): Promise<ConsciousnessState> => {
    const response = await api.get(`/consciousness/state/${userId}`);
    return response.data.consciousness_state;
  },

  // 个性化适配
  adaptPersonality: async (request: AdaptationRequest) => {
    const response = await api.post('/consciousness/adapt', request);
    return response.data;
  },

  // 获取融合历史
  getFusionHistory: async (userId: number, limit?: number): Promise<FusionHistory[]> => {
    const params = limit ? { limit } : {};
    const response = await api.get(`/consciousness/history/${userId}`, { params });
    return response.data.history;
  },

  // 意识状态实时监控
  subscribeToStateUpdates: (userId: number, callback: (state: ConsciousnessState) => void) => {
    // WebSocket连接实现
    const wsUrl = process.env.REACT_APP_WS_URL || 'ws://localhost:8080';
    const ws = new WebSocket(`${wsUrl}/consciousness/subscribe/${userId}`);
    
    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.type === 'consciousness_update') {
        callback(data.state);
      }
    };
    
    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };
    
    ws.onclose = () => {
      console.log('WebSocket connection closed');
    };
    
    return () => ws.close();
  },

  // 获取意识类型列表
  getConsciousnessTypes: async () => {
    return [
      { value: 'beginner', label: '初学者', description: '刚开始意识修炼的阶段' },
      { value: 'intermediate', label: '进阶者', description: '有一定修炼基础的阶段' },
      { value: 'advanced', label: '高级者', description: '深度修炼的阶段' },
      { value: 'master', label: '大师级', description: '达到高度觉悟的阶段' },
    ];
  },

  // 获取目标状态列表
  getTargetStates: async () => {
    return [
      { value: 'calm', label: '平静', description: '达到内心平静的状态' },
      { value: 'focused', label: '专注', description: '提高注意力和专注度' },
      { value: 'enlightened', label: '开悟', description: '获得更高层次的觉悟' },
      { value: 'balanced', label: '平衡', description: '身心灵的和谐平衡' },
      { value: 'compassionate', label: '慈悲', description: '培养慈悲心和同理心' },
    ];
  },

  // 获取分析类型列表
  getAnalysisTypes: async () => {
    return [
      { value: 'emotional', label: '情绪分析', description: '分析当前的情绪状态' },
      { value: 'cognitive', label: '认知分析', description: '评估认知能力和思维模式' },
      { value: 'spiritual', label: '精神分析', description: '探索精神层面的状态' },
      { value: 'comprehensive', label: '综合分析', description: '全面的意识状态评估' },
    ];
  },

  // 创建默认的意识状态
  createDefaultState: (userId: number): Partial<ConsciousnessState> => {
    return {
      user_id: userId,
      emotional_state: {
        calm: 0.5,
        happy: 0.5,
        focused: 0.5,
        peaceful: 0.5,
      },
      cognitive_level: 0.5,
      spiritual_depth: 0.5,
      personality_traits: {
        openness: 0.5,
        conscientiousness: 0.5,
        extraversion: 0.5,
        agreeableness: 0.5,
        neuroticism: 0.5,
      },
      consciousness_type: 'beginner',
      fusion_level: 0.0,
      metadata: {},
    };
  },

  // 计算意识状态得分
  calculateConsciousnessScore: (state: ConsciousnessState): number => {
    const emotionalAvg = Object.values(state.emotional_state).reduce((a, b) => a + b, 0) / Object.values(state.emotional_state).length;
    const personalityAvg = Object.values(state.personality_traits).reduce((a, b) => a + b, 0) / Object.values(state.personality_traits).length;
    
    return Math.round(
      (emotionalAvg * 0.3 + 
       state.cognitive_level * 0.25 + 
       state.spiritual_depth * 0.25 + 
       personalityAvg * 0.1 + 
       state.fusion_level * 0.1) * 100
    );
  },

  // 生成意识提升建议
  generateSuggestions: (state: ConsciousnessState): string[] => {
    const suggestions: string[] = [];
    
    if (state.cognitive_level < 0.6) {
      suggestions.push('建议进行专注力训练，如冥想或正念练习');
    }
    
    if (state.spiritual_depth < 0.6) {
      suggestions.push('尝试深度冥想或精神修炼，提升精神层次');
    }
    
    if (state.fusion_level < 0.3) {
      suggestions.push('参与更多的意识融合练习，提高融合水平');
    }
    
    const emotionalAvg = Object.values(state.emotional_state).reduce((a, b) => a + b, 0) / Object.values(state.emotional_state).length;
    if (emotionalAvg < 0.6) {
      suggestions.push('关注情绪调节，保持内心平静和积极心态');
    }
    
    if (suggestions.length === 0) {
      suggestions.push('继续保持当前的修炼状态，稳步提升');
    }
    
    return suggestions;
  },
};

// 导出类型
export type {
  ConsciousnessState,
  FusionRequest,
  AnalysisRequest,
  AdaptationRequest,
  FusionHistory,
};