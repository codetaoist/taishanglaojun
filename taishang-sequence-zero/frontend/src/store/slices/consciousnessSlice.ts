import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { message } from 'antd';
import { consciousnessService } from '../../services/consciousnessService';
import type {
  ConsciousnessState as ConsciousnessStateType,
  FusionRequest,
  AnalysisRequest,
  AdaptationRequest,
  FusionHistory
} from '../../services/consciousnessService';

// 意识实体接口
export interface ConsciousnessEntity {
  id: string;
  name: string;
  type: 'human' | 'ai' | 'hybrid';
  level: number;
  state: string;
  energy: number;
  wisdom: number;
  harmony: number;
  lastActive: string;
  attributes: Record<string, any>;
  connections: string[];
}

// 融合会话接口
export interface FusionSession {
  id: string;
  participants: string[];
  mode: string;
  startTime: string;
  endTime?: string;
  status: 'preparing' | 'active' | 'completed' | 'failed';
  progress: number;
  insights: string[];
  energyFlow: number[];
  harmonyIndex: number;
}

// 意识洞察接口
export interface ConsciousnessInsight {
  id: string;
  title: string;
  content: string;
  category: 'wisdom' | 'emotion' | 'logic' | 'intuition';
  confidence: number;
  timestamp: string;
  source: string;
  tags: string[];
}

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

// Redux状态接口
export interface ConsciousnessSliceState {
  // 当前用户意识实体
  currentEntity: ConsciousnessEntity | null;
  
  // 当前意识状态
  currentState: ConsciousnessStateType | null;
  
  // 所有意识实体
  entities: ConsciousnessEntity[];
  
  // 当前融合会话
  currentSession: FusionSession | null;
  
  // 历史会话
  sessions: FusionSession[];
  
  // 融合历史
  fusionHistory: FusionHistory[];
  
  // 意识洞察
  insights: ConsciousnessInsight[];
  
  // 分析结果
  analysisResults: any[];
  
  // 适配建议
  adaptationSuggestions: any[];
  
  // 加载状态
  loading: boolean;
  sessionLoading: boolean;
  analysisLoading: boolean;
  fusionLoading: boolean;
  
  // 错误信息
  error: string | null;
  
  // 统计数据
  stats: {
    totalSessions: number;
    successfulFusions: number;
    averageHarmony: number;
    totalInsights: number;
    energyLevel: number;
    wisdomGrowth: number;
    consciousnessScore: number;
  };
  
  // 设置
  settings: {
    autoFusion: boolean;
    fusionMode: string;
    notificationEnabled: boolean;
    privacyLevel: 'open' | 'selective' | 'private';
    energyThreshold: number;
    autoAnalysis: boolean;
  };
  
  // WebSocket连接状态
  wsConnected: boolean;
}

// 初始状态
const initialState: ConsciousnessSliceState = {
  currentEntity: null,
  currentState: null,
  entities: [],
  currentSession: null,
  sessions: [],
  fusionHistory: [],
  insights: [],
  analysisResults: [],
  adaptationSuggestions: [],
  loading: false,
  sessionLoading: false,
  analysisLoading: false,
  fusionLoading: false,
  error: null,
  stats: {
    totalSessions: 0,
    successfulFusions: 0,
    averageHarmony: 0,
    totalInsights: 0,
    energyLevel: 100,
    wisdomGrowth: 0,
    consciousnessScore: 0,
  },
  settings: {
    autoFusion: false,
    fusionMode: 'gentle',
    notificationEnabled: true,
    privacyLevel: 'selective',
    energyThreshold: 50,
    autoAnalysis: false,
  },
  wsConnected: false,
};

// 异步thunk：获取当前意识状态
export const fetchConsciousnessState = createAsyncThunk(
  'consciousness/fetchState',
  async (_, { rejectWithValue }) => {
    try {
      const response = await consciousnessService.getConsciousnessState(1); // 默认用户ID
      return response;
    } catch (error: any) {
      return rejectWithValue(error.message || '获取意识状态失败');
    }
  }
);

// 异步thunk：分析意识
export const analyzeConsciousness = createAsyncThunk(
  'consciousness/analyze',
  async (request: AnalysisRequest, { rejectWithValue }) => {
    try {
      const response = await consciousnessService.analyzeConsciousness(request);
      if (response.success) {
        return response.data;
      } else {
        return rejectWithValue(response.message || '意识分析失败');
      }
    } catch (error: any) {
      return rejectWithValue(error.message || '意识分析失败');
    }
  }
);

// 异步thunk：融合意识
export const fuseConsciousness = createAsyncThunk(
  'consciousness/fuse',
  async (request: FusionRequest, { rejectWithValue }) => {
    try {
      const response = await consciousnessService.fuseConsciousness(request);
      if (response.success) {
        message.success('意识融合成功');
        return response.data;
      } else {
        return rejectWithValue(response.message || '意识融合失败');
      }
    } catch (error: any) {
      return rejectWithValue(error.message || '意识融合失败');
    }
  }
);

// 异步thunk：适配个性
export const adaptPersonality = createAsyncThunk(
  'consciousness/adapt',
  async (request: AdaptationRequest, { rejectWithValue }) => {
    try {
      const response = await consciousnessService.adaptPersonality(request);
      if (response.success) {
        message.success('个性适配成功');
        return response.data;
      } else {
        return rejectWithValue(response.message || '个性适配失败');
      }
    } catch (error: any) {
      return rejectWithValue(error.message || '个性适配失败');
    }
  }
);

// 异步thunk：获取融合历史
export const fetchFusionHistory = createAsyncThunk(
  'consciousness/fetchHistory',
  async (params: { page?: number; limit?: number }, { rejectWithValue }) => {
    try {
      const response = await consciousnessService.getFusionHistory(params.page || 1, params.limit || 10);
      return response;
    } catch (error: any) {
      return rejectWithValue(error.message || '获取融合历史失败');
    }
  }
);

// 异步thunk：获取意识类型列表
export const fetchConsciousnessTypes = createAsyncThunk(
  'consciousness/fetchTypes',
  async (_, { rejectWithValue }) => {
    try {
      const response = await consciousnessService.getConsciousnessTypes();
      return response;
    } catch (error: any) {
      return rejectWithValue(error.message || '获取意识类型失败');
    }
  }
);

// 异步thunk：获取目标状态列表
export const fetchTargetStates = createAsyncThunk(
  'consciousness/fetchTargetStates',
  async (_, { rejectWithValue }) => {
    try {
      const response = await consciousnessService.getTargetStates();
      return response;
    } catch (error: any) {
      return rejectWithValue(error.message || '获取目标状态失败');
    }
  }
);

// 异步thunk：获取分析类型列表
export const fetchAnalysisTypes = createAsyncThunk(
  'consciousness/fetchAnalysisTypes',
  async (_, { rejectWithValue }) => {
    try {
      const response = await consciousnessService.getAnalysisTypes();
      return response;
    } catch (error: any) {
      return rejectWithValue(error.message || '获取分析类型失败');
    }
  }
);

// 异步thunk：生成建议
export const generateSuggestions = createAsyncThunk(
  'consciousness/generateSuggestions',
  async (state: ConsciousnessStateType, { rejectWithValue }) => {
    try {
      const response = await consciousnessService.generateSuggestions(state);
      return response;
    } catch (error: any) {
      return rejectWithValue(error.message || '生成建议失败');
    }
  }
);

// 异步thunk：初始化意识
export const initializeConsciousness = createAsyncThunk(
  'consciousness/initialize',
  async (userId: number, { rejectWithValue }) => {
    try {
      const response = await consciousnessService.getConsciousnessState(userId);
      return response;
    } catch (error: any) {
      return rejectWithValue(error.message || '初始化意识失败');
    }
  }
);

// 异步thunk：获取洞察
export const fetchInsights = createAsyncThunk(
  'consciousness/fetchInsights',
  async (userId: number, { rejectWithValue }) => {
    try {
      const response = await consciousnessService.getFusionHistory(userId, 10);
      // 将历史记录转换为洞察格式
      const insights = response.map((history, index) => ({
        id: `insight-${history.id}`,
        title: `融合会话洞察 #${history.id}`,
        content: history.feedback || '无具体反馈',
        category: 'wisdom' as const,
        confidence: history.effectiveness,
        timestamp: history.start_time,
        source: 'fusion-history',
        tags: ['融合', '历史'],
      }));
      return insights;
    } catch (error: any) {
      return rejectWithValue(error.message || '获取洞察失败');
    }
  }
);

// 异步thunk：开始融合会话
export const startFusionSession = createAsyncThunk(
  'consciousness/startSession',
  async (request: { participants: string[]; mode: string }, { rejectWithValue }) => {
    try {
      const fusionRequest: FusionRequest = {
        user_id: 1, // 默认用户ID
        target_state: 'balanced',
        intensity: 0.5,
        duration: 30,
        personalization: {
          mode: request.mode,
          participants: request.participants,
        },
      };
      const response = await consciousnessService.fuseConsciousness(fusionRequest);
      return {
        session_id: `session-${Date.now()}`,
        participants: request.participants,
        fusion_type: request.mode,
        harmony_score: 0.5,
        ...response,
      };
    } catch (error: any) {
      return rejectWithValue(error.message || '开始融合会话失败');
    }
  }
);

// 异步thunk：结束融合会话
export const endFusionSession = createAsyncThunk(
  'consciousness/endSession',
  async (sessionId: string, { rejectWithValue }) => {
    try {
      // 这里可以调用后端API来结束会话
      // 暂时返回模拟数据
      return {
        sessionId,
        endTime: new Date().toISOString(),
        status: 'completed' as const,
        summary: '融合会话已成功完成',
      };
    } catch (error: any) {
      return rejectWithValue(error.message || '结束融合会话失败');
    }
  }
);

// 创建意识slice
const consciousnessSlice = createSlice({
  name: 'consciousness',
  initialState,
  reducers: {
    // 清除错误
    clearError: (state) => {
      state.error = null;
    },
    
    // 设置加载状态
    setLoading: (state, action: PayloadAction<boolean>) => {
      state.loading = action.payload;
    },
    
    // 设置分析加载状态
    setAnalysisLoading: (state, action: PayloadAction<boolean>) => {
      state.analysisLoading = action.payload;
    },
    
    // 设置融合加载状态
    setFusionLoading: (state, action: PayloadAction<boolean>) => {
      state.fusionLoading = action.payload;
    },
    
    // 更新当前实体
    updateCurrentEntity: (state, action: PayloadAction<Partial<ConsciousnessEntity>>) => {
      if (state.currentEntity) {
        state.currentEntity = { ...state.currentEntity, ...action.payload };
      }
    },
    
    // 更新当前状态
    updateCurrentState: (state, action: PayloadAction<Partial<ConsciousnessStateType>>) => {
      if (state.currentState) {
        state.currentState = { ...state.currentState, ...action.payload };
      }
    },
    
    // 添加洞察
    addInsight: (state, action: PayloadAction<ConsciousnessInsight>) => {
      state.insights.unshift(action.payload);
      state.stats.totalInsights += 1;
    },
    
    // 更新会话进度
    updateSessionProgress: (state, action: PayloadAction<{ sessionId: string; progress: number }>) => {
      if (state.currentSession && state.currentSession.id === action.payload.sessionId) {
        state.currentSession.progress = action.payload.progress;
      }
    },
    
    // 更新能量水平
    updateEnergyLevel: (state, action: PayloadAction<number>) => {
      state.stats.energyLevel = Math.max(0, Math.min(100, action.payload));
      if (state.currentEntity) {
        state.currentEntity.energy = state.stats.energyLevel;
      }
      if (state.currentState) {
        state.currentState.cognitive_level = state.stats.energyLevel;
      }
    },
    
    // 更新和谐指数
    updateHarmonyIndex: (state, action: PayloadAction<number>) => {
      if (state.currentSession) {
        state.currentSession.harmonyIndex = action.payload;
      }
      if (state.currentState) {
        state.currentState.spiritual_depth = action.payload;
      }
    },
    
    // 更新意识分数
    updateConsciousnessScore: (state, action: PayloadAction<number>) => {
      state.stats.consciousnessScore = action.payload;
    },
    
    // 更新设置
    updateSettings: (state, action: PayloadAction<Partial<ConsciousnessSliceState['settings']>>) => {
      state.settings = { ...state.settings, ...action.payload };
    },
    
    // 设置WebSocket连接状态
    setWsConnected: (state, action: PayloadAction<boolean>) => {
      state.wsConnected = action.payload;
    },
    
    // WebSocket状态更新
    onStateUpdate: (state, action: PayloadAction<ConsciousnessStateType>) => {
      state.currentState = action.payload;
      // 更新统计数据
      const score = consciousnessService.calculateConsciousnessScore(action.payload);
      state.stats.consciousnessScore = score;
    },
    
    // 重置状态
    resetConsciousnessState: () => initialState,
  },
  extraReducers: (builder) => {
    // 获取意识状态
    builder
      .addCase(fetchConsciousnessState.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchConsciousnessState.fulfilled, (state, action) => {
        state.loading = false;
        state.currentState = action.payload;
        
        // 实体相关逻辑暂时移除
        
        // 更新统计数据
        const score = consciousnessService.calculateConsciousnessScore(action.payload);
        state.stats.consciousnessScore = score;
        state.stats.energyLevel = action.payload.cognitive_level;
        
        state.error = null;
      })
      .addCase(fetchConsciousnessState.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
        message.error(action.payload as string);
      });
    
    // 分析意识
    builder
      .addCase(analyzeConsciousness.pending, (state) => {
        state.analysisLoading = true;
        state.error = null;
      })
      .addCase(analyzeConsciousness.fulfilled, (state, action) => {
        state.analysisLoading = false;
        state.analysisResults.unshift(action.payload);
        state.error = null;
        message.success('意识分析完成');
      })
      .addCase(analyzeConsciousness.rejected, (state, action) => {
        state.analysisLoading = false;
        state.error = action.payload as string;
        message.error(action.payload as string);
      });
    
    // 融合意识
    builder
      .addCase(fuseConsciousness.pending, (state) => {
        state.fusionLoading = true;
        state.error = null;
      })
      .addCase(fuseConsciousness.fulfilled, (state, action) => {
        state.fusionLoading = false;
        
        // 创建融合会话
        const session: FusionSession = {
          id: action.payload.session_id,
          participants: action.payload.participants || [],
          mode: action.payload.fusion_type,
          startTime: new Date().toISOString(),
          status: 'active',
          progress: 0,
          insights: [],
          energyFlow: [],
          harmonyIndex: action.payload.harmony_score || 0.5,
        };
        
        state.currentSession = session;
        state.stats.totalSessions += 1;
        
        if (action.payload.success) {
          state.stats.successfulFusions += 1;
        }
        
        state.error = null;
        message.success('意识融合成功');
      })
      .addCase(fuseConsciousness.rejected, (state, action) => {
        state.fusionLoading = false;
        state.error = action.payload as string;
        message.error(action.payload as string);
      });
    
    // 适配个性
    builder
      .addCase(adaptPersonality.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(adaptPersonality.fulfilled, (state, action) => {
        state.loading = false;
        state.adaptationSuggestions.unshift(action.payload);
        state.error = null;
        message.success('个性适配完成');
      })
      .addCase(adaptPersonality.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
        message.error(action.payload as string);
      });
    
    // 获取融合历史
    builder
      .addCase(fetchFusionHistory.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchFusionHistory.fulfilled, (state, action) => {
        state.loading = false;
        state.fusionHistory = action.payload;
        state.error = null;
      })
      .addCase(fetchFusionHistory.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      });
    
    // 获取意识类型
    builder
      .addCase(fetchConsciousnessTypes.fulfilled, (state, action) => {
        // 可以存储到单独的状态中，这里暂时不处理
      });
    
    // 获取目标状态
    builder
      .addCase(fetchTargetStates.fulfilled, (state, action) => {
        // 可以存储到单独的状态中，这里暂时不处理
      });
    
    // 获取分析类型
    builder
      .addCase(fetchAnalysisTypes.fulfilled, (state, action) => {
        // 可以存储到单独的状态中，这里暂时不处理
      });
    
    // 生成建议
    builder
      .addCase(generateSuggestions.fulfilled, (state, action) => {
        const suggestions = action.payload.map((suggestion: any, index: number) => ({
          id: `suggestion-${Date.now()}-${index}`,
          title: suggestion.title || '意识建议',
          content: suggestion.description || suggestion.content,
          category: 'wisdom' as const,
          confidence: suggestion.confidence || 0.8,
          timestamp: new Date().toISOString(),
          source: 'consciousness-service',
          tags: suggestion.tags || ['建议'],
        }));
        
        state.insights = [...suggestions, ...state.insights];
        state.stats.totalInsights += suggestions.length;
      });
    
    // 初始化意识
    builder
      .addCase(initializeConsciousness.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(initializeConsciousness.fulfilled, (state, action) => {
        state.loading = false;
        state.currentState = action.payload;
        state.error = null;
      })
      .addCase(initializeConsciousness.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      });
    
    // 获取洞察
    builder
      .addCase(fetchInsights.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchInsights.fulfilled, (state, action) => {
        state.loading = false;
        state.insights = action.payload;
        state.stats.totalInsights = action.payload.length;
        state.error = null;
      })
      .addCase(fetchInsights.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      });
    
    // 开始融合会话
    builder
      .addCase(startFusionSession.pending, (state) => {
        state.sessionLoading = true;
        state.error = null;
      })
      .addCase(startFusionSession.fulfilled, (state, action) => {
        state.sessionLoading = false;
        const session: FusionSession = {
          id: action.payload.session_id,
          participants: action.payload.participants || [],
          mode: action.payload.fusion_type,
          startTime: new Date().toISOString(),
          status: 'active',
          progress: 0,
          insights: [],
          energyFlow: [],
          harmonyIndex: action.payload.harmony_score || 0.5,
        };
        state.currentSession = session;
        state.sessions.unshift(session);
        state.stats.totalSessions += 1;
        state.error = null;
      })
      .addCase(startFusionSession.rejected, (state, action) => {
        state.sessionLoading = false;
        state.error = action.payload as string;
      });
    
    // 结束融合会话
    builder
      .addCase(endFusionSession.pending, (state) => {
        state.sessionLoading = true;
        state.error = null;
      })
      .addCase(endFusionSession.fulfilled, (state, action) => {
        state.sessionLoading = false;
        if (state.currentSession) {
          state.currentSession.status = 'completed';
          state.currentSession.endTime = action.payload.endTime;
          state.currentSession.progress = 100;
          state.stats.successfulFusions += 1;
        }
        state.currentSession = null;
        state.error = null;
      })
      .addCase(endFusionSession.rejected, (state, action) => {
        state.sessionLoading = false;
        state.error = action.payload as string;
      });
  },
});

// 导出actions
export const {
  clearError,
  setLoading,
  setAnalysisLoading,
  setFusionLoading,
  updateCurrentEntity,
  updateCurrentState,
  addInsight,
  updateSessionProgress,
  updateEnergyLevel,
  updateHarmonyIndex,
  updateConsciousnessScore,
  updateSettings,
  setWsConnected,
  onStateUpdate,
  resetConsciousnessState,
} = consciousnessSlice.actions;

// 导出异步thunk函数
export {
  fetchConsciousnessState,
  analyzeConsciousness,
  fuseConsciousness,
  adaptPersonality,
  fetchFusionHistory,
  fetchConsciousnessTypes,
  fetchTargetStates,
  fetchAnalysisTypes,
  generateSuggestions,
  initializeConsciousness,
  fetchInsights,
  startFusionSession,
  endFusionSession,
};

// 选择器
export const selectConsciousness = (state: { consciousness: ConsciousnessSliceState }) => state.consciousness;
export const selectCurrentEntity = (state: { consciousness: ConsciousnessSliceState }) => state.consciousness.currentEntity;
export const selectCurrentState = (state: { consciousness: ConsciousnessSliceState }) => state.consciousness.currentState;
export const selectCurrentSession = (state: { consciousness: ConsciousnessSliceState }) => state.consciousness.currentSession;
export const selectInsights = (state: { consciousness: ConsciousnessSliceState }) => state.consciousness.insights;
export const selectConsciousnessStats = (state: { consciousness: ConsciousnessSliceState }) => state.consciousness.stats;
export const selectConsciousnessSettings = (state: { consciousness: ConsciousnessSliceState }) => state.consciousness.settings;
export const selectFusionHistory = (state: { consciousness: ConsciousnessSliceState }) => state.consciousness.fusionHistory;
export const selectAnalysisResults = (state: { consciousness: ConsciousnessSliceState }) => state.consciousness.analysisResults;
export const selectAdaptationSuggestions = (state: { consciousness: ConsciousnessSliceState }) => state.consciousness.adaptationSuggestions;
export const selectWsConnected = (state: { consciousness: ConsciousnessSliceState }) => state.consciousness.wsConnected;
export const selectConsciousnessLoading = (state: { consciousness: ConsciousnessSliceState }) => ({
  general: state.consciousness.loading,
  analysis: state.consciousness.analysisLoading,
  fusion: state.consciousness.fusionLoading,
  session: state.consciousness.sessionLoading,
});

// 导出reducer
export default consciousnessSlice.reducer;