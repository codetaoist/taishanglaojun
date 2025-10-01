import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { message } from 'antd';
import { culturalService } from '../../services/culturalService';
// 文化条目接口
export interface CulturalEntry {
  id: string;
  title: string;
  content: string;
  category: CulturalCategory;
  tags: string[];
  author?: string;
  source?: string;
  createdAt: string;
  updatedAt: string;
  viewCount: number;
  likeCount: number;
}

// 文化问题接口
export interface CulturalQuestion {
  id: string;
  question: string;
  options: string[];
  correctAnswer: number;
  explanation: string;
  category: CulturalCategory;
  difficulty: number;
  tags: string[];
  source?: string;
}

// 文化测验接口
export interface CulturalQuiz {
  id: string;
  title: string;
  description: string;
  category: CulturalCategory;
  level: WisdomLevel;
  questions: CulturalQuestion[];
  timeLimit?: number;
  passingScore: number;
  tags: string[];
}

// 文化进度接口
export interface CulturalProgress {
  id: string;
  userId: string;
  category: CulturalCategory;
  level: WisdomLevel;
  completedLessons: number;
  totalLessons: number;
  practiceHours: number;
  achievements: string[];
  lastActivity: string;
  currentStreak: number;
  maxStreak: number;
}

// 文化分析接口
export interface CulturalAnalysis {
  id: string;
  content: string;
  category: CulturalCategory;
  insights: string[];
  recommendations: string[];
  culturalElements: string[];
  historicalContext: string;
  modernRelevance: string;
  practicalApplications: string[];
  relatedConcepts: string[];
}

// 文化推荐接口
export interface CulturalRecommendation {
  id: string;
  title: string;
  description: string;
  category: CulturalCategory;
  level: WisdomLevel;
  reason: string;
  priority: number;
  estimatedTime: number;
  prerequisites?: string[];
  relatedItems: string[];
}

// 文化搜索参数接口
export interface CulturalSearchParams {
  query?: string;
  category?: CulturalCategory;
  level?: WisdomLevel;
  tags?: string[];
  limit?: number;
  offset?: number;
}

// 文化类别类型
export type CulturalCategory = 'philosophy' | 'literature' | 'history' | 'art' | 'medicine' | 'astronomy' | 'ethics';

// 智慧等级类型
export type WisdomLevel = 'beginner' | 'intermediate' | 'advanced' | 'master' | 'sage';

// 文化智慧条目接口
export interface WisdomItem {
  id: string;
  title: string;
  content: string;
  category: CulturalCategory;
  level: WisdomLevel;
  source: string;
  author?: string;
  dynasty?: string;
  tags: string[];
  difficulty: number;
  popularity: number;
  createdAt: string;
  updatedAt: string;
}

// 学习进度接口
export interface LearningProgress {
  itemId: string;
  progress: number;
  completed: boolean;
  startTime: string;
  completedTime?: string;
  notes: string[];
  bookmarked: boolean;
}

// 智慧对话接口
export interface WisdomDialogue {
  id: string;
  question: string;
  answer: string;
  category: CulturalCategory;
  confidence: number;
  sources: string[];
  timestamp: string;
  helpful: boolean | null;
}





// 测验结果接口
export interface QuizResult {
  quizId: string;
  score: number;
  totalQuestions: number;
  correctAnswers: number;
  timeSpent: number;
  completedAt: string;
  answers: Array<{
    questionId: string;
    selectedAnswer: number;
    correct: boolean;
  }>;
}

// 文化状态接口
export interface CulturalSliceState {
  // 智慧条目
  wisdomItems: WisdomItem[];
  currentItem: WisdomItem | null;
  
  // 学习进度
  learningProgress: Record<string, LearningProgress>;
  
  // 用户整体进度
  userProgress: any;
  
  // 对话历史
  dialogues: WisdomDialogue[];
  currentDialogue: WisdomDialogue | null;
  
  // 测验相关
  quizzes: CulturalQuiz[];
  currentQuiz: CulturalQuiz | null;
  quizResults: QuizResult[];
  
  // 搜索和筛选
  searchQuery: string;
  selectedCategory: CulturalCategory | 'all';
  selectedLevel: WisdomLevel | 'all';
  
  // 加载状态
  loading: boolean;
  dialogueLoading: boolean;
  quizLoading: boolean;
  
  // 错误信息
  error: string | null;
  
  // 统计数据
  stats: {
    totalItems: number;
    completedItems: number;
    totalDialogues: number;
    averageScore: number;
    studyTime: number;
    wisdomLevel: WisdomLevel;
    achievements: string[];
  };
  
  // 用户偏好
  preferences: {
    favoriteCategories: CulturalCategory[];
    studyReminder: boolean;
    difficultyPreference: WisdomLevel;
    displayMode: 'card' | 'list' | 'timeline';
    fontSize: 'small' | 'medium' | 'large';
  };
  
  // 收藏和书签
  bookmarks: string[];
  favorites: string[];
  
  // 推荐内容
  recommendations: any[];
  
  // 文化类别
  categories: any[];
  
  // 当前分析结果
  currentAnalysis: any;
}

// 初始状态
const initialState: CulturalSliceState = {
  wisdomItems: [],
  currentItem: null,
  learningProgress: {},
  userProgress: null,
  dialogues: [],
  currentDialogue: null,
  quizzes: [],
  currentQuiz: null,
  quizResults: [],
  searchQuery: '',
  selectedCategory: 'all',
  selectedLevel: 'all',
  loading: false,
  dialogueLoading: false,
  quizLoading: false,
  error: null,
  stats: {
    totalItems: 0,
    completedItems: 0,
    totalDialogues: 0,
    averageScore: 0,
    studyTime: 0,
    wisdomLevel: 'beginner',
    achievements: [],
  },
  preferences: {
    favoriteCategories: [],
    studyReminder: true,
    difficultyPreference: 'intermediate',
    displayMode: 'card',
    fontSize: 'medium',
  },
  bookmarks: [],
  favorites: [],
  recommendations: [],
  categories: [],
  currentAnalysis: null,
};

// 异步thunk：获取文化条目
export const fetchCulturalEntries = createAsyncThunk(
  'cultural/fetchEntries',
  async (params: CulturalSearchParams, { rejectWithValue }) => {
    try {
      const response = await culturalService.searchCulturalContent(params.query || '', params.category);
      if (response.success) {
        return response.data;
      } else {
        return rejectWithValue(response.message || '获取文化条目失败');
      }
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || '获取文化条目失败';
      return rejectWithValue(errorMessage);
    }
  }
);

// 异步thunk：提问文化问题
export const askCulturalQuestion = createAsyncThunk(
  'cultural/askQuestion',
  async (question: string, { rejectWithValue }) => {
    try {
      const response = await culturalService.inquireWisdom({
        question: question,
        category: 'philosophy'
      });
      if (response.success) {
        return response.data;
      } else {
        const errorMessage = response.message || '文化问答失败';
        message.error(errorMessage);
        return rejectWithValue(errorMessage);
      }
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || '文化问答失败';
      message.error(errorMessage);
      return rejectWithValue(errorMessage);
    }
  }
);

// 异步thunk：获取文化测验
export const fetchCulturalQuizzes = createAsyncThunk(
  'cultural/fetchQuizzes',
  async (params: { category?: CulturalCategory; difficulty?: string }, { rejectWithValue }) => {
    try {
      const response = await culturalService.getWisdomQAs(params.category, 10);
      return response;
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || '获取测验失败';
      return rejectWithValue(errorMessage);
    }
  }
);

// 异步thunk：提交测验答案
export const submitCulturalQuiz = createAsyncThunk(
  'cultural/submitQuiz',
  async (params: { quizId: string; answers: Record<string, string> }, { rejectWithValue }) => {
    try {
      const response = await culturalService.recordPractice({
        practice_type: 'cultural_quiz',
        duration_minutes: 15,
        intensity: 4,
        notes: `完成测验ID: ${params.quizId}`,
        insights: [`提交答案: ${JSON.stringify(params.answers)}`],
        effectiveness_rating: 5,
        emotional_state_before: {},
        emotional_state_after: {}
      });
      if (response.success && response.data) {
        message.success(`测验完成！得分：${response.data.score}`);
        return response.data;
      } else {
        const errorMessage = response.message || '提交测验失败';
        message.error(errorMessage);
        return rejectWithValue(errorMessage);
      }
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || '提交测验失败';
      message.error(errorMessage);
      return rejectWithValue(errorMessage);
    }
  }
);

// 异步thunk：更新学习进度
export const updateCulturalProgress = createAsyncThunk(
  'cultural/updateProgress',
  async (params: { entryId: string; progress: number; completed?: boolean }, { rejectWithValue }) => {
    try {
      const response = await culturalService.recordPractice({
        practice_type: 'cultural_study',
        duration_minutes: 30,
        intensity: 3,
        notes: `学习进度更新: ${params.progress}%`,
        insights: [`完成状态: ${params.completed ? '已完成' : '进行中'}`],
        effectiveness_rating: 4,
        emotional_state_before: {},
        emotional_state_after: {}
      });
      if (response.success) {
        return response.data;
      } else {
        return rejectWithValue(response.message || '更新进度失败');
      }
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || '更新进度失败';
      return rejectWithValue(errorMessage);
    }
  }
);

// 异步thunk：获取文化分析
export const getCulturalAnalysis = createAsyncThunk(
  'cultural/getAnalysis',
  async (entryId: string, { rejectWithValue }) => {
    try {
      const response = await culturalService.inquireWisdom({
        question: `请分析条目ID为${entryId}的文化内容`,
        category: 'philosophy',
        context: '文化内容分析'
      });
      if (response.success) {
        return response.data;
      } else {
        return rejectWithValue(response.message || '获取文化分析失败');
      }
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || '获取文化分析失败';
      return rejectWithValue(errorMessage);
    }
  }
);

// 异步thunk：获取文化推荐
export const getCulturalRecommendations = createAsyncThunk(
  'cultural/getRecommendations',
  async (userId: string, { rejectWithValue }) => {
    try {
      const response = await culturalService.generateCultivationSuggestions(Number(userId), {});
      if (response.success) {
        return response.data;
      } else {
        return rejectWithValue(response.message || '获取文化推荐失败');
      }
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || '获取文化推荐失败';
      return rejectWithValue(errorMessage);
    }
  }
);

// 异步thunk：获取文化类别
export const fetchCulturalCategories = createAsyncThunk(
  'cultural/fetchCategories',
  async (_, { rejectWithValue }) => {
    try {
      const response = await culturalService.getCulturalCategories();
      return response;
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || '获取文化类别失败';
      return rejectWithValue(errorMessage);
    }
  }
);

// 异步thunk：获取用户进度
export const fetchUserProgress = createAsyncThunk(
  'cultural/fetchUserProgress',
  async (userId: string, { rejectWithValue }) => {
    try {
      const response = await culturalService.getCultivationStats(Number(userId));
      if (response.success) {
        return response.data;
      } else {
        return rejectWithValue(response.message || '获取用户进度失败');
      }
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || '获取用户进度失败';
      return rejectWithValue(errorMessage);
    }
  }
);

// 创建文化slice
const culturalSlice = createSlice({
  name: 'cultural',
  initialState,
  reducers: {
    // 清除错误
    clearError: (state) => {
      state.error = null;
    },
    
    // 设置当前智慧条目
    setCurrentItem: (state, action: PayloadAction<WisdomItem | null>) => {
      state.currentItem = action.payload;
    },
    
    // 设置搜索查询
    setSearchQuery: (state, action: PayloadAction<string>) => {
      state.searchQuery = action.payload;
    },
    
    // 设置选中的类别
    setSelectedCategory: (state, action: PayloadAction<CulturalCategory | 'all'>) => {
      state.selectedCategory = action.payload;
    },
    
    // 设置选中的等级
    setSelectedLevel: (state, action: PayloadAction<WisdomLevel | 'all'>) => {
      state.selectedLevel = action.payload;
    },
    
    // 添加书签
    addBookmark: (state, action: PayloadAction<string>) => {
      if (!state.bookmarks.includes(action.payload)) {
        state.bookmarks.push(action.payload);
      }
    },
    
    // 移除书签
    removeBookmark: (state, action: PayloadAction<string>) => {
      state.bookmarks = state.bookmarks.filter(id => id !== action.payload);
    },
    
    // 添加收藏
    addFavorite: (state, action: PayloadAction<string>) => {
      if (!state.favorites.includes(action.payload)) {
        state.favorites.push(action.payload);
      }
    },
    
    // 移除收藏
    removeFavorite: (state, action: PayloadAction<string>) => {
      state.favorites = state.favorites.filter(id => id !== action.payload);
    },
    
    // 评价对话
    rateDialogue: (state, action: PayloadAction<{ dialogueId: string; helpful: boolean }>) => {
      const dialogue = state.dialogues.find(d => d.id === action.payload.dialogueId);
      if (dialogue) {
        dialogue.helpful = action.payload.helpful;
      }
      if (state.currentDialogue && state.currentDialogue.id === action.payload.dialogueId) {
        state.currentDialogue.helpful = action.payload.helpful;
      }
    },
    
    // 更新偏好设置
    updatePreferences: (state, action: PayloadAction<Partial<CulturalSliceState['preferences']>>) => {
      state.preferences = { ...state.preferences, ...action.payload };
    },
    
    // 添加学习笔记
    addLearningNote: (state, action: PayloadAction<{ itemId: string; note: string }>) => {
      const progress = state.learningProgress[action.payload.itemId];
      if (progress) {
        progress.notes.push(action.payload.note);
      }
    },
    
    // 清除对话历史
    clearDialogueHistory: (state) => {
      state.dialogues = [];
      state.currentDialogue = null;
    },
    
    // 重置文化状态
    resetCulturalState: () => initialState,
  },
  extraReducers: (builder) => {
    // 获取文化条目
    builder
      .addCase(fetchCulturalEntries.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchCulturalEntries.fulfilled, (state, action) => {
        state.loading = false;
        if (action.payload.success && action.payload.data) {
           state.wisdomItems = action.payload.data.map((item: any) => ({
             ...item,
             category: item.category as CulturalCategory || 'philosophy',
             level: item.difficulty as WisdomLevel || 'beginner',
             difficulty: 1, // 默认难度值
             popularity: item.viewCount || 0 // 使用viewCount作为popularity
           }));
         }
        if (action.payload.pagination) {
          state.stats.totalItems = action.payload.pagination.total;
        }
      })
      .addCase(fetchCulturalEntries.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      })
      
    // 文化问答
    builder
      .addCase(askCulturalQuestion.pending, (state) => {
        state.dialogueLoading = true;
        state.error = null;
      })
      .addCase(askCulturalQuestion.fulfilled, (state, action) => {
        state.dialogueLoading = false;
        if (action.payload.success && action.payload.data) {
          const dialogueData: WisdomDialogue = {
            id: action.payload.data.id || Date.now().toString(),
            question: action.payload.data.question,
            answer: action.payload.data.answer,
            category: action.payload.data.category as CulturalCategory || 'philosophy',
            confidence: action.payload.data.confidence || 0.8,
            sources: action.payload.data.sources || [],
            timestamp: new Date().toISOString(),
            helpful: null
          };
          state.currentDialogue = dialogueData;
          state.dialogues.unshift(dialogueData);
        }
        state.stats.totalDialogues += 1;
      })
      .addCase(askCulturalQuestion.rejected, (state, action) => {
        state.dialogueLoading = false;
        state.error = action.payload as string;
      })
      
    // 获取测验
    builder
      .addCase(fetchCulturalQuizzes.pending, (state) => {
        state.quizLoading = true;
        state.error = null;
      })
      .addCase(fetchCulturalQuizzes.fulfilled, (state, action) => {
        state.quizLoading = false;
        state.quizzes = action.payload.map((quiz: any) => ({
          ...quiz,
          category: quiz.category as CulturalCategory || 'philosophy',
          level: quiz.difficulty as WisdomLevel || 'beginner',
          questions: quiz.questions || [],
          id: quiz.id || Math.random().toString()
        }));
      })
      .addCase(fetchCulturalQuizzes.rejected, (state, action) => {
        state.quizLoading = false;
        state.error = action.payload as string;
      })
      
    // 提交测验
    builder
      .addCase(submitCulturalQuiz.fulfilled, (state, action) => {
        if (action.payload.success && action.payload.data) {
          const quizResult: QuizResult = {
            quizId: action.meta.arg.quizId,
            score: action.payload.data.score,
            totalQuestions: action.payload.data.totalQuestions || 0,
            correctAnswers: action.payload.data.correctAnswers || 0,
            timeSpent: action.payload.data.timeSpent || 0,
            completedAt: new Date().toISOString(),
            answers: action.payload.data.answers || []
          };
          state.quizResults.unshift(quizResult);
          // 更新平均分数
          const totalScore = state.quizResults.reduce((sum, result) => sum + result.score, 0);
          state.stats.averageScore = totalScore / state.quizResults.length;
        }
      })
      
    // 更新学习进度
    builder
      .addCase(updateCulturalProgress.fulfilled, (state, action) => {
        if (action.payload.success && action.payload.data) {
          const progressData = action.payload.data;
          state.learningProgress[progressData.entryId] = {
            itemId: progressData.entryId,
            progress: progressData.progress,
            completed: progressData.completed,
            startTime: progressData.startedAt,
            completedTime: progressData.completedAt,
            notes: progressData.notes ? [progressData.notes] : [],
            bookmarked: false
          };
          // 更新完成数量
          const completedCount = Object.values(state.learningProgress).filter(p => p.completed).length;
          state.stats.completedItems = completedCount;
        }
      })
      
    // 获取文化分析
    builder
      .addCase(getCulturalAnalysis.fulfilled, (state, action) => {
        if (action.payload.success && action.payload.data) {
          // 可以将分析结果存储到state中
          state.currentAnalysis = action.payload.data;
        }
      })
      
    // 获取文化推荐
    builder
      .addCase(getCulturalRecommendations.fulfilled, (state, action) => {
        if (action.payload.success && action.payload.data) {
          state.recommendations = action.payload.data;
        }
      })
      
    // 获取文化类别
    builder
      .addCase(fetchCulturalCategories.fulfilled, (state, action) => {
        state.categories = action.payload;
      })
      
    // 获取用户进度
    builder
      .addCase(fetchUserProgress.fulfilled, (state, action) => {
        if (action.payload.success && action.payload.data) {
          // 更新用户的整体学习进度
          state.userProgress = action.payload.data;
          // 更新统计信息
          if (action.payload.data.stats) {
            state.stats = { ...state.stats, ...action.payload.data.stats };
          }
        }
      });
  },
});

// 导出actions
export const {
  clearError,
  setCurrentItem,
  setSearchQuery,
  setSelectedCategory,
  setSelectedLevel,
  addBookmark,
  removeBookmark,
  addFavorite,
  removeFavorite,
  rateDialogue,
  updatePreferences,
  addLearningNote,
  clearDialogueHistory,
  resetCulturalState,
} = culturalSlice.actions;

// 导出异步thunk函数
export {
  fetchCulturalEntries as fetchWisdomItems,
  askCulturalQuestion as startWisdomDialogue,
  fetchCulturalQuizzes,
  submitQuizAnswer,
  updateCulturalProgress,
  getCulturalAnalysis,
  getCulturalRecommendations,
  fetchCulturalCategories,
  fetchUserProgress,
};

// 选择器
export const selectCultural = (state: { cultural: CulturalSliceState }) => state.cultural;
export const selectWisdomItems = (state: { cultural: CulturalSliceState }) => state.cultural.wisdomItems;
export const selectCurrentItem = (state: { cultural: CulturalSliceState }) => state.cultural.currentItem;
export const selectDialogues = (state: { cultural: CulturalSliceState }) => state.cultural.dialogues;
export const selectCulturalStats = (state: { cultural: CulturalSliceState }) => state.cultural.stats;
export const selectBookmarks = (state: { cultural: CulturalSliceState }) => state.cultural.bookmarks;
export const selectFavorites = (state: { cultural: CulturalSliceState }) => state.cultural.favorites;
export const selectCulturalLoading = (state: { cultural: CulturalSliceState }) => state.cultural.loading;
export const selectCulturalError = (state: { cultural: CulturalSliceState }) => state.cultural.error;

// 导出reducer
export default culturalSlice.reducer;