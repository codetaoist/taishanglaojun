import { api } from './api';
import type { ApiResponse } from './api';

// 模型接口
export interface Model {
  id: string;
  name: string;
  type: string;
  provider: string;
  version?: string;
  description?: string;
  config?: any;
  capabilities?: string[];
  createdAt: string;
  updatedAt: string;
}

// 向量集合相关接口
export interface Collection {
  id: number;
  tenantId: string;
  name: string;
  description?: string;
  modelId?: string;
  dims: number;
  indexType: string;
  metricType: string;
  extraIndexArgs?: any;
  createdAt: string;
  updatedAt: string;
}

// 向量集合列表响应接口
export interface CollectionListResponse {
  total: number;
  page: number;
  pageSize: number;
  items: Collection[];
}

// 任务状态枚举
export enum TaskStatus {
  Pending = 'pending',
  Running = 'running',
  Completed = 'completed',
  Failed = 'failed',
  Cancelled = 'cancelled'
}

// 任务类型枚举
export enum TaskType {
  Indexing = 'indexing',
  Training = 'training',
  Inference = 'inference',
  FineTuning = 'fine_tuning',
  DataProcessing = 'data_processing'
}

// 任务接口
export interface Task {
  id: string;
  name: string;
  type: TaskType;
  status: TaskStatus;
  description?: string;
  config?: any;
  input?: any;
  output?: any;
  progress?: number;
  error?: string;
  createdAt: string;
  updatedAt: string;
  startedAt?: string;
  completedAt?: string;
}

// 模型管理API
export const modelApi = {
  // 获取所有模型
  getAll: (): Promise<ApiResponse<Model[]>> => {
    return api.get<Model[]>('/api/taishang/models');
  },
  
  // 获取单个模型
  get: (id: string): Promise<ApiResponse<Model>> => {
    return api.get<Model>(`/api/taishang/models/${id}`);
  },
  
  // 注册模型
  register: (model: Omit<Model, 'id' | 'createdAt' | 'updatedAt'>): Promise<ApiResponse<Model>> => {
    return api.post<Model>('/api/taishang/models', model);
  },
  
  // 更新模型
  update: (id: string, model: Partial<Model>): Promise<ApiResponse<Model>> => {
    return api.put<Model>(`/api/taishang/models/${id}`, model);
  },
  
  // 删除模型
  delete: (id: string): Promise<ApiResponse<void>> => {
    return api.delete<void>(`/api/taishang/models/${id}`);
  }
};

// 向量集合管理API
export const collectionApi = {
  // 获取所有集合
  getAll: (): Promise<ApiResponse<CollectionListResponse>> => {
    return api.get<CollectionListResponse>('/api/v1/vector/collections');
  },
  
  // 获取单个集合
  get: (id: number): Promise<ApiResponse<Collection>> => {
    return api.get<Collection>(`/api/v1/vector/collections/${id}`);
  },
  
  // 创建集合
  create: (collection: Omit<Collection, 'id' | 'createdAt' | 'updatedAt'>): Promise<ApiResponse<Collection>> => {
    return api.post<Collection>('/api/v1/vector/collections', collection);
  },
  
  // 更新集合
  update: (id: number, collection: Partial<Collection>): Promise<ApiResponse<Collection>> => {
    return api.put<Collection>(`/api/v1/vector/collections/${id}`, collection);
  },
  
  // 删除集合
  delete: (id: number): Promise<ApiResponse<void>> => {
    return api.delete<void>(`/api/v1/vector/collections/${id}`);
  },
  
  // 重建集合索引
  rebuildIndex: (id: number): Promise<ApiResponse<void>> => {
    return api.post<void>(`/api/v1/vector/collections/${id}/rebuild-index`);
  }
};

// 任务管理API
export const taskApi = {
  // 获取所有任务
  getAll: (status?: TaskStatus, type?: TaskType): Promise<ApiResponse<Task[]>> => {
    const params = new URLSearchParams();
    if (status) params.append('status', status);
    if (type) params.append('type', type);
    
    const queryString = params.toString();
    const url = queryString ? `/api/taishang/tasks?${queryString}` : '/api/taishang/tasks';
    
    return api.get<Task[]>(url);
  },
  
  // 获取单个任务
  get: (id: string): Promise<ApiResponse<Task>> => {
    return api.get<Task>(`/api/taishang/tasks/${id}`);
  },
  
  // 创建任务
  create: (task: Omit<Task, 'id' | 'createdAt' | 'updatedAt' | 'startedAt' | 'completedAt'>): Promise<ApiResponse<Task>> => {
    return api.post<Task>('/api/taishang/tasks', task);
  },
  
  // 更新任务
  update: (id: string, task: Partial<Task>): Promise<ApiResponse<Task>> => {
    return api.put<Task>(`/api/taishang/tasks/${id}`, task);
  },
  
  // 删除任务
  delete: (id: string): Promise<ApiResponse<void>> => {
    return api.delete<void>(`/api/taishang/tasks/${id}`);
  }
};