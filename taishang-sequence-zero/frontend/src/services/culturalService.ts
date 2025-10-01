import { api } from './api';

// 智慧问答接口
export interface WisdomQA {
  id: number;
  question: string;
  answer: string;
  category: string;
  tags: string[];
  difficulty_level: number;
  cultural_context: string;
  source: string;
  created_at: string;
  updated_at: string;
}

// 文化知识接口
export interface CulturalKnowledge {
  id: number;
  title: string;
  content: string;
  category: string;
  tags: string[];
  knowledge_type: string;
  difficulty_level: number;
  cultural_period: string;
  region: string;
  related_concepts: string[];
  multimedia_resources: Record<string, any>;
  created_at: string;
  updated_at: string;
}

// 修炼计划接口
export interface CultivationPlan {
  id: number;
  user_id: number;
  plan_name: string;
  description: string;
  goals: string[];
  practices: Record<string, any>;
  duration_days: number;
  difficulty_level: number;
  cultural_focus: string;
  progress_tracking: Record<string, any>;
  created_at: string;
  updated_at: string;
}

// 修炼记录接口
export interface PracticeRecord {
  id: number;
  user_id: number;
  plan_id?: number;
  practice_type: string;
  duration_minutes: number;
  intensity: number;
  notes: string;
  insights: string[];
  effectiveness_rating: number;
  emotional_state_before: Record<string, number>;
  emotional_state_after: Record<string, number>;
  practice_date: string;
  created_at: string;
}

// 文化传承故事接口
export interface HeritageStory {
  id: number;
  title: string;
  content: string;
  story_type: string;
  cultural_period: string;
  region: string;
  characters: string[];
  moral_lessons: string[];
  cultural_significance: string;
  related_practices: string[];
  multimedia_content: Record<string, any>;
  created_at: string;
  updated_at: string;
}

// 文化体验接口
export interface CulturalExperience {
  id: number;
  user_id: number;
  experience_type: string;
  title: string;
  description: string;
  cultural_elements: string[];
  learning_outcomes: string[];
  reflection: string;
  rating: number;
  duration_minutes: number;
  experience_date: string;
  created_at: string;
}

// 智慧问答请求接口
export interface WisdomInquiryRequest {
  question: string;
  category?: string;
  context?: string;
  user_level?: number;
}

// 文化知识查询请求接口
export interface KnowledgeQueryRequest {
  query: string;
  category?: string;
  knowledge_type?: string;
  difficulty_level?: number;
  cultural_period?: string;
  region?: string;
}

// 修炼计划创建请求接口
export interface CreateCultivationPlanRequest {
  plan_name: string;
  description: string;
  goals: string[];
  duration_days: number;
  difficulty_level: number;
  cultural_focus: string;
  preferred_practices: string[];
}

// 修炼记录创建请求接口
export interface CreatePracticeRecordRequest {
  plan_id?: number;
  practice_type: string;
  duration_minutes: number;
  intensity: number;
  notes: string;
  insights: string[];
  effectiveness_rating: number;
  emotional_state_before: Record<string, number>;
  emotional_state_after: Record<string, number>;
}

// 文化体验创建请求接口
export interface CreateCulturalExperienceRequest {
  experience_type: string;
  title: string;
  description: string;
  cultural_elements: string[];
  learning_outcomes: string[];
  reflection: string;
  rating: number;
  duration_minutes: number;
}

// 文化智慧服务
export const culturalService = {
  // 智慧问答
  inquireWisdom: async (request: WisdomInquiryRequest) => {
    const response = await api.post('/cultural/wisdom/inquire', request);
    return response.data;
  },

  // 获取智慧问答列表
  getWisdomQAs: async (category?: string, limit?: number) => {
    const params = { category, limit };
    const response = await api.get('/cultural/wisdom/qa', { params });
    return response.data.qa_list as WisdomQA[];
  },

  // 搜索文化知识
  searchKnowledge: async (request: KnowledgeQueryRequest) => {
    const response = await api.post('/cultural/knowledge/search', request);
    return response.data;
  },

  // 获取文化知识详情
  getKnowledgeDetail: async (id: number): Promise<CulturalKnowledge> => {
    const response = await api.get(`/cultural/knowledge/${id}`);
    return response.data.knowledge;
  },

  // 获取文化知识列表
  getKnowledgeList: async (category?: string, limit?: number) => {
    const params = { category, limit };
    const response = await api.get('/cultural/knowledge', { params });
    return response.data.knowledge_list as CulturalKnowledge[];
  },

  // 创建修炼计划
  createCultivationPlan: async (request: CreateCultivationPlanRequest) => {
    const response = await api.post('/cultural/cultivation/plan', request);
    return response.data;
  },

  // 获取用户修炼计划
  getUserCultivationPlans: async (userId: number): Promise<CultivationPlan[]> => {
    const response = await api.get(`/cultural/cultivation/plans/${userId}`);
    return response.data.plans;
  },

  // 更新修炼计划
  updateCultivationPlan: async (planId: number, updates: Partial<CultivationPlan>) => {
    const response = await api.put(`/cultural/cultivation/plan/${planId}`, updates);
    return response.data;
  },

  // 删除修炼计划
  deleteCultivationPlan: async (planId: number) => {
    const response = await api.delete(`/cultural/cultivation/plan/${planId}`);
    return response.data;
  },

  // 记录修炼实践
  recordPractice: async (request: CreatePracticeRecordRequest) => {
    const response = await api.post('/cultural/cultivation/record', request);
    return response.data;
  },

  // 获取修炼记录
  getPracticeRecords: async (userId: number, planId?: number, limit?: number): Promise<PracticeRecord[]> => {
    const params = { plan_id: planId, limit };
    const response = await api.get(`/cultural/cultivation/records/${userId}`, { params });
    return response.data.records;
  },

  // 获取文化传承故事
  getHeritageStories: async (storyType?: string, region?: string, limit?: number): Promise<HeritageStory[]> => {
    const params = { story_type: storyType, region, limit };
    const response = await api.get('/cultural/heritage/stories', { params });
    return response.data.stories;
  },

  // 获取故事详情
  getStoryDetail: async (id: number): Promise<HeritageStory> => {
    const response = await api.get(`/cultural/heritage/story/${id}`);
    return response.data.story;
  },

  // 创建文化体验记录
  createCulturalExperience: async (request: CreateCulturalExperienceRequest) => {
    const response = await api.post('/cultural/experience', request);
    return response.data;
  },

  // 获取用户文化体验
  getUserExperiences: async (userId: number, experienceType?: string, limit?: number): Promise<CulturalExperience[]> => {
    const params = { experience_type: experienceType, limit };
    const response = await api.get(`/cultural/experiences/${userId}`, { params });
    return response.data.experiences;
  },

  // 获取文化类别列表
  getCulturalCategories: async () => {
    return [
      { value: 'philosophy', label: '哲学思想', description: '中华哲学智慧与思辨' },
      { value: 'meditation', label: '冥想修炼', description: '传统冥想与静心方法' },
      { value: 'ethics', label: '道德伦理', description: '传统道德观念与伦理准则' },
      { value: 'literature', label: '文学经典', description: '古典文学与诗词歌赋' },
      { value: 'medicine', label: '中医养生', description: '传统医学与养生之道' },
      { value: 'martial_arts', label: '武术功法', description: '传统武术与内功修炼' },
      { value: 'tea_culture', label: '茶道文化', description: '茶艺与茶道精神' },
      { value: 'calligraphy', label: '书法艺术', description: '书法修炼与艺术美学' },
      { value: 'music', label: '古典音乐', description: '传统音乐与音律之美' },
      { value: 'architecture', label: '建筑文化', description: '传统建筑与空间美学' },
    ];
  },

  // 获取修炼类型列表
  getPracticeTypes: async () => {
    return [
      { value: 'meditation', label: '静坐冥想', description: '通过静坐达到心灵宁静' },
      { value: 'breathing', label: '呼吸调息', description: '调节呼吸，平衡身心' },
      { value: 'qigong', label: '气功练习', description: '调动内气，强身健体' },
      { value: 'taichi', label: '太极拳', description: '柔和缓慢的武术修炼' },
      { value: 'reading', label: '经典诵读', description: '诵读经典，启发智慧' },
      { value: 'calligraphy', label: '书法练习', description: '通过书写修炼心性' },
      { value: 'tea_ceremony', label: '茶道修习', description: '品茶悟道，陶冶情操' },
      { value: 'nature_walk', label: '自然漫步', description: '亲近自然，感悟天地' },
      { value: 'music_practice', label: '音乐修炼', description: '通过音乐净化心灵' },
      { value: 'contemplation', label: '哲理思辨', description: '深度思考人生哲理' },
    ];
  },

  // 获取文化时期列表
  getCulturalPeriods: async () => {
    return [
      { value: 'pre_qin', label: '先秦时期', description: '春秋战国，百家争鸣' },
      { value: 'qin_han', label: '秦汉时期', description: '大一统时代的文化奠基' },
      { value: 'wei_jin', label: '魏晋南北朝', description: '玄学兴起，文化多元' },
      { value: 'sui_tang', label: '隋唐时期', description: '盛世文化，开放包容' },
      { value: 'song_yuan', label: '宋元时期', description: '理学发展，文化精深' },
      { value: 'ming_qing', label: '明清时期', description: '传统文化的集大成' },
      { value: 'modern', label: '近现代', description: '传统与现代的融合' },
    ];
  },

  // 获取地区列表
  getRegions: async () => {
    return [
      { value: 'central_plains', label: '中原地区', description: '华夏文明的发源地' },
      { value: 'jiangnan', label: '江南地区', description: '文人雅士的聚集地' },
      { value: 'sichuan', label: '巴蜀地区', description: '独特的巴蜀文化' },
      { value: 'lingnan', label: '岭南地区', description: '开放包容的南方文化' },
      { value: 'northeast', label: '东北地区', description: '豪放质朴的北方文化' },
      { value: 'northwest', label: '西北地区', description: '丝路文化的交汇点' },
      { value: 'tibet', label: '西藏地区', description: '神秘深邃的雪域文化' },
      { value: 'xinjiang', label: '新疆地区', description: '多民族文化的融合' },
    ];
  },

  // 获取难度等级列表
  getDifficultyLevels: async () => {
    return [
      { value: 1, label: '入门级', description: '适合初学者，基础概念' },
      { value: 2, label: '初级', description: '有一定基础，简单应用' },
      { value: 3, label: '中级', description: '需要一定经验，深入理解' },
      { value: 4, label: '高级', description: '需要丰富经验，复杂应用' },
      { value: 5, label: '专家级', description: '需要深厚功底，精深研究' },
    ];
  },

  // 生成个性化修炼建议
  generateCultivationSuggestions: async (userId: number, preferences: Record<string, any>) => {
    const response = await api.post('/cultural/cultivation/suggestions', {
      user_id: userId,
      preferences,
    });
    return response.data.suggestions;
  },

  // 获取修炼进度统计
  getCultivationStats: async (userId: number) => {
    const response = await api.get(`/cultural/cultivation/stats/${userId}`);
    return response.data.stats;
  },

  // 获取文化智慧日报
  getDailyWisdom: async () => {
    const response = await api.get('/cultural/daily-wisdom');
    return response.data;
  },

  // 搜索相关文化内容
  searchCulturalContent: async (query: string, contentType?: string) => {
    const params = { q: query, type: contentType };
    const response = await api.get('/cultural/search', { params });
    return response.data;
  },
};

// 导出类型
export type {
  WisdomQA,
  CulturalKnowledge,
  CultivationPlan,
  PracticeRecord,
  HeritageStory,
  CulturalExperience,
  WisdomInquiryRequest,
  KnowledgeQueryRequest,
  CreateCultivationPlanRequest,
  CreatePracticeRecordRequest,
  CreateCulturalExperienceRequest,
};