import axios from 'axios';
import type { AxiosInstance, AxiosResponse } from 'axios';

// 智能学习相关类型定义
export interface LearnerProfile {
  id: string;
  userId: string;
  name: string;
  email: string;
  avatar?: string;
  level: string;
  experience: number;
  nextLevelExp: number;
  learningGoals: string[];
  preferences: LearningPreferences;
  skills: Skill[];
  achievements: Achievement[];
  createdAt: string;
  updatedAt: string;
}

export interface LearningPreferences {
  learningStyle: 'visual' | 'auditory' | 'kinesthetic' | 'reading';
  difficulty: 'beginner' | 'intermediate' | 'advanced';
  topics: string[];
  studyTime: number; // 每日学习时间（分钟）
  reminderEnabled: boolean;
  reminderTime: string;
}

export interface Skill {
  id: string;
  name: string;
  category: string;
  level: number; // 0-100
  progress: number; // 0-100
  lastPracticed: string;
}

export interface Achievement {
  id: string;
  title: string;
  description: string;
  icon: string;
  rarity: 'common' | 'rare' | 'epic' | 'legendary';
  earnedAt: string;
  progress?: number;
  maxProgress?: number;
}

export interface LearningContent {
  id: string;
  title: string;
  description: string;
  type: 'course' | 'lesson' | 'exercise' | 'assessment';
  category: string;
  difficulty: 'beginner' | 'intermediate' | 'advanced';
  duration: number; // 分钟
  thumbnail: string;
  tags: string[];
  rating: number;
  reviewCount: number;
  author: {
    id: string;
    name: string;
    avatar?: string;
  };
  prerequisites: string[];
  learningObjectives: string[];
  content: ContentSection[];
  createdAt: string;
  updatedAt: string;
}

export interface ContentSection {
  id: string;
  title: string;
  type: 'text' | 'video' | 'audio' | 'interactive' | 'quiz';
  content: string;
  duration?: number;
  resources?: Resource[];
}

export interface Resource {
  id: string;
  title: string;
  type: 'pdf' | 'video' | 'audio' | 'link' | 'image';
  url: string;
  size?: number;
}

export interface LearningPath {
  id: string;
  title: string;
  description: string;
  difficulty: 'beginner' | 'intermediate' | 'advanced';
  estimatedDuration: number; // 小时
  modules: LearningModule[];
  prerequisites: string[];
  skills: string[];
  progress: number; // 0-100
  status: 'not_started' | 'in_progress' | 'completed';
  createdAt: string;
  updatedAt: string;
}

export interface LearningModule {
  id: string;
  title: string;
  description: string;
  order: number;
  content: LearningContent[];
  progress: number; // 0-100
  status: 'locked' | 'available' | 'in_progress' | 'completed';
  estimatedDuration: number; // 分钟
}

export interface LearningProgress {
  id: string;
  learnerId: string;
  contentId: string;
  pathId?: string;
  moduleId?: string;
  progress: number; // 0-100
  status: 'not_started' | 'in_progress' | 'completed' | 'paused';
  timeSpent: number; // 分钟
  lastAccessed: string;
  completedAt?: string;
  score?: number;
  notes?: string;
}

export interface LearningAnalytics {
  learnerId: string;
  totalStudyTime: number; // 小时
  coursesCompleted: number;
  currentStreak: number; // 连续学习天数
  averageScore: number;
  skillProgress: SkillProgress[];
  weeklyActivity: ActivityData[];
  monthlyProgress: ProgressData[];
  recommendations: Recommendation[];
}

export interface SkillProgress {
  skillId: string;
  skillName: string;
  currentLevel: number;
  progress: number;
  trend: 'up' | 'down' | 'stable';
}

export interface ActivityData {
  date: string;
  studyTime: number; // 分钟
  coursesCompleted: number;
  exercisesCompleted: number;
}

export interface ProgressData {
  month: string;
  coursesCompleted: number;
  skillsImproved: number;
  averageScore: number;
}

export interface Recommendation {
  id: string;
  type: 'course' | 'exercise' | 'assessment' | 'skill_practice';
  title: string;
  description: string;
  reason: string;
  confidence: number; // 0-100
  priority: 'low' | 'medium' | 'high';
  estimatedTime: number; // 分钟
  difficulty: 'beginner' | 'intermediate' | 'advanced';
  tags: string[];
  thumbnail?: string;
}

export interface KnowledgeGraphNode {
  id: string;
  title: string;
  type: 'concept' | 'skill' | 'topic' | 'course';
  description: string;
  level: number;
  prerequisites: string[];
  relatedNodes: string[];
  mastery: number; // 0-100
}

export interface LearningSession {
  id: string;
  learnerId: string;
  contentId: string;
  startTime: string;
  endTime?: string;
  duration: number; // 分钟
  progress: number; // 0-100
  interactions: SessionInteraction[];
  performance: SessionPerformance;
}

export interface SessionInteraction {
  timestamp: string;
  type: 'view' | 'click' | 'pause' | 'resume' | 'complete' | 'quiz_answer';
  data: Record<string, any>;
}

export interface SessionPerformance {
  accuracy: number; // 0-100
  speed: number; // 相对速度
  engagement: number; // 0-100
  difficulty: number; // 感知难度 0-100
}

// API响应类型
interface ApiResponse<T = unknown> {
  success: boolean;
  data: T;
  message?: string;
  error?: string;
  pagination?: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
}

// 查询参数类型
export interface QueryParams {
  page?: number;
  limit?: number;
  search?: string;
  category?: string;
  difficulty?: string;
  tags?: string[];
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
}

export interface LearningQueryParams extends QueryParams {
  type?: string;
  status?: string;
  minRating?: number;
  maxDuration?: number;
}

class LearningApiClient {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: import.meta.env.VITE_LEARNING_API_URL || 'http://localhost:8080/api/v1',
      timeout: 15000,
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
        if (error.response?.status === 401) {
          localStorage.removeItem('token');
          // 可以触发全局状态更新或导航到登录页
        }
        return Promise.reject(error);
      }
    );
  }

  // 学习者管理API
  async getLearnerProfile(learnerId?: string): Promise<ApiResponse<LearnerProfile>> {
    const url = learnerId ? `/learners/${learnerId}` : '/learners/profile';
    const response = await this.client.get(url);
    return response.data;
  }

  // 获取周活动数据
  async getWeeklyActivity(): Promise<ApiResponse<ActivityData[]>> {
    const response = await this.client.get('/learning/weekly-activity');
    return response.data;
  }

  async updateLearnerProfile(learnerId: string, data: Partial<LearnerProfile>): Promise<ApiResponse<LearnerProfile>> {
    const response = await this.client.put(`/learners/${learnerId}`, data);
    return response.data;
  }

  async updateLearningPreferences(learnerId: string, preferences: LearningPreferences): Promise<ApiResponse<LearningPreferences>> {
    const response = await this.client.put(`/learners/${learnerId}/preferences`, preferences);
    return response.data;
  }

  // 学习内容API
  async getLearningContent(params?: LearningQueryParams): Promise<ApiResponse<LearningContent[]>> {
    const response = await this.client.get('/content', { params });
    return response.data;
  }

  async getContentById(contentId: string): Promise<ApiResponse<LearningContent>> {
    const response = await this.client.get(`/content/${contentId}`);
    return response.data;
  }

  async searchContent(query: string, params?: LearningQueryParams): Promise<ApiResponse<LearningContent[]>> {
    const response = await this.client.get('/content/search', { 
      params: { q: query, ...params } 
    });
    return response.data;
  }

  // 学习路径API
  async getLearningPaths(params?: QueryParams): Promise<ApiResponse<LearningPath[]>> {
    const response = await this.client.get('/learning-paths', { params });
    return response.data;
  }

  async getLearningPathById(pathId: string): Promise<ApiResponse<LearningPath>> {
    const response = await this.client.get(`/learning-paths/${pathId}`);
    return response.data;
  }

  async enrollInLearningPath(pathId: string): Promise<ApiResponse<void>> {
    const response = await this.client.post(`/learning-paths/${pathId}/enroll`);
    return response.data;
  }

  async getPersonalizedLearningPath(learnerId: string, goals: string[]): Promise<ApiResponse<LearningPath>> {
    const response = await this.client.post(`/learners/${learnerId}/personalized-path`, { goals });
    return response.data;
  }

  // 学习进度API
  async getLearningProgress(learnerId: string, params?: QueryParams): Promise<ApiResponse<LearningProgress[]>> {
    const response = await this.client.get(`/learners/${learnerId}/progress`, { params });
    return response.data;
  }

  async updateLearningProgress(progressId: string, data: Partial<LearningProgress>): Promise<ApiResponse<LearningProgress>> {
    const response = await this.client.put(`/progress/${progressId}`, data);
    return response.data;
  }

  async startLearningSession(contentId: string): Promise<ApiResponse<LearningSession>> {
    const response = await this.client.post('/sessions', { contentId });
    return response.data;
  }

  async endLearningSession(sessionId: string, data: Partial<LearningSession>): Promise<ApiResponse<LearningSession>> {
    const response = await this.client.put(`/sessions/${sessionId}/end`, data);
    return response.data;
  }

  // 学习分析API
  async getLearningAnalytics(learnerId: string, timeRange?: string): Promise<ApiResponse<LearningAnalytics>> {
    const response = await this.client.get('/learners/analytics', {
      params: { timeRange }
    });
    return response.data;
  }

  async getSkillProgress(learnerId: string): Promise<ApiResponse<SkillProgress[]>> {
    const response = await this.client.get('/learning/skill-progress');
    return response.data;
  }

  async getRecommendations(learnerId: string, type?: string): Promise<ApiResponse<Recommendation[]>> {
    const response = await this.client.get('/learning/recommendations', {
      params: { type }
    });
    return response.data;
  }

  // 知识图谱API
  async getKnowledgeGraph(params?: QueryParams): Promise<ApiResponse<KnowledgeGraphNode[]>> {
    const response = await this.client.get('/knowledge-graph', { params });
    return response.data;
  }

  async getKnowledgeGraphNode(nodeId: string): Promise<ApiResponse<KnowledgeGraphNode>> {
    const response = await this.client.get(`/knowledge-graph/nodes/${nodeId}`);
    return response.data;
  }

  async getRelatedNodes(nodeId: string): Promise<ApiResponse<KnowledgeGraphNode[]>> {
    const response = await this.client.get(`/knowledge-graph/nodes/${nodeId}/related`);
    return response.data;
  }

  // 实时分析API
  async getRealtimeAnalytics(learnerId: string): Promise<ApiResponse<any>> {
    const response = await this.client.get(`/learners/${learnerId}/realtime-analytics`);
    return response.data;
  }

  async trackLearningBehavior(learnerId: string, behavior: SessionInteraction): Promise<ApiResponse<void>> {
    const response = await this.client.post(`/learners/${learnerId}/track-behavior`, behavior);
    return response.data;
  }

  // 成就系统API
  async getAchievements(learnerId: string): Promise<ApiResponse<Achievement[]>> {
    const response = await this.client.get('/learning/achievements');
    return response.data;
  }

  async unlockAchievement(learnerId: string, achievementId: string): Promise<ApiResponse<Achievement>> {
    const response = await this.client.post(`/learners/${learnerId}/achievements/${achievementId}/unlock`);
    return response.data;
  }
}

// 创建单例实例
export const learningApi = new LearningApiClient();
export default learningApi;