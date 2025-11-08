import { api } from './api';

// 业务API相关的类型定义
export interface Plugin {
  id: string;
  name: string;
  version: string;
  status: 'running' | 'stopped' | 'error';
  description?: string;
  checksum?: string;
  created_at?: string;
  updated_at?: string;
}

export interface PluginListResponse {
  total: number;
  page: number;
  pageSize: number;
  items: Plugin[];
}

export interface PluginInstallRequest {
  pluginId: string;
  version: string;
  sourceUrl?: string;
}

export interface PluginActionRequest {
  pluginId: string;
}

export interface PluginActionResponse {
  success: boolean;
  pluginId: string;
  message?: string;
}

// 审计日志相关类型
export interface AuditLog {
  id: number;
  actor: string;
  action: string;
  resource: string;
  status: 'SUCCESS' | 'FAILED';
  message: string;
  created_at: string;
}

export interface AuditListResponse {
  total: number;
  page: number;
  pageSize: number;
  items: AuditLog[];
}

// 配置相关类型
export interface Config {
  key: string;
  value: string;
  scope: 'global' | 'domain' | 'plugin';
  description?: string;
}

export interface ConfigListResponse {
  total: number;
  items: Config[];
}

export interface ConfigUpdateRequest {
  value: string;
  scope: 'global' | 'domain' | 'plugin';
}

// 模型相关类型
export interface Model {
  id: string;
  name: string;
  version: string;
  family?: string;
  quantization?: string;
  status: 'enabled' | 'disabled';
  params?: Record<string, any>;
  created_at?: string;
}

export interface ModelListResponse {
  total: number;
  page: number;
  pageSize: number;
  items: Model[];
}

export interface ModelCreateRequest {
  name: string;
  version: string;
  family?: string;
  quantization?: string;
  params?: Record<string, any>;
}

// 向量集合相关类型
export interface VectorCollection {
  id: string;
  name: string;
  dim: number;
  indexType: 'HNSW' | 'IVF' | 'FLAT';
  metric: 'cosine' | 'l2' | 'dot';
  replication?: number;
  created_at?: string;
}

export interface VectorCollectionListResponse {
  total: number;
  page: number;
  pageSize: number;
  items: VectorCollection[];
}

export interface VectorCollectionCreateRequest {
  name: string;
  dim: number;
  indexType: 'HNSW' | 'IVF' | 'FLAT';
  metric: 'cosine' | 'l2' | 'dot';
  replication?: number;
}

export interface VectorUpsertRequest {
  namespace?: string;
  vectors: Array<{
    id: string;
    values: number[];
    metadata?: Record<string, any>;
  }>;
}

export interface VectorUpsertResponse {
  upserted: number;
  namespace: string;
}

export interface VectorQueryRequest {
  topK: number;
  query: {
    text?: string;
    string?: string;
    embedding?: number[];
  };
  namespace?: string;
}

export interface VectorQueryResponse {
  matches: Array<{
    id: string;
    score: number;
    metadata?: Record<string, any>;
  }>;
  count: number;
}

// 任务相关类型
export interface Task {
  id: string;
  type: string;
  status: 'PENDING' | 'RUNNING' | 'SUCCEEDED' | 'FAILED' | 'CANCELED';
  priority: 'low' | 'normal' | 'high';
  payload: Record<string, any>;
  result?: Record<string, any>;
  created_at?: string;
  updated_at?: string;
}

export interface TaskListResponse {
  total: number;
  page: number;
  pageSize: number;
  items: Task[];
}

export interface TaskCreateRequest {
  type: string;
  payload: Record<string, any>;
  priority?: 'low' | 'normal' | 'high';
  ttl?: number;
}

// 业务API服务
export const businessApi = {
  // 插件管理
  plugins: {
    // 获取插件列表
    list: async (params?: {
      status?: string;
      name?: string;
      page?: number;
      pageSize?: number;
    }): Promise<PluginListResponse> => {
      const queryParams = new URLSearchParams();
      if (params?.status) queryParams.append('status', params.status);
      if (params?.name) queryParams.append('name', params.name);
      if (params?.page) queryParams.append('page', params.page.toString());
      if (params?.pageSize) queryParams.append('pageSize', params.pageSize.toString());
      
      const query = queryParams.toString();
      return api.get<PluginListResponse>(`/api/laojun/plugins/list${query ? `?${query}` : ''}`);
    },

    // 安装插件
    install: async (data: PluginInstallRequest): Promise<PluginActionResponse> => {
      return api.post<PluginActionResponse>('/api/laojun/plugins/install', data);
    },

    // 启动插件
    start: async (data: PluginActionRequest): Promise<PluginActionResponse> => {
      return api.post<PluginActionResponse>('/api/laojun/plugins/start', data);
    },

    // 停止插件
    stop: async (data: PluginActionRequest): Promise<PluginActionResponse> => {
      return api.post<PluginActionResponse>('/api/laojun/plugins/stop', data);
    },

    // 升级插件
    upgrade: async (data: { pluginId: string; version: string }): Promise<PluginActionResponse> => {
      return api.post<PluginActionResponse>('/api/laojun/plugins/upgrade', data);
    },

    // 卸载插件
    uninstall: async (data: PluginActionRequest): Promise<PluginActionResponse> => {
      return api.delete<PluginActionResponse>('/api/laojun/plugins/uninstall', { data });
    },
  },

  // 审计日志
  audits: {
    // 获取审计日志列表
    list: async (params?: {
      actor?: string;
      action?: string;
      resource?: string;
      from?: string;
      to?: string;
      page?: number;
      pageSize?: number;
    }): Promise<AuditListResponse> => {
      const queryParams = new URLSearchParams();
      if (params?.actor) queryParams.append('actor', params.actor);
      if (params?.action) queryParams.append('action', params.action);
      if (params?.resource) queryParams.append('resource', params.resource);
      if (params?.from) queryParams.append('from', params.from);
      if (params?.to) queryParams.append('to', params.to);
      if (params?.page) queryParams.append('page', params.page.toString());
      if (params?.pageSize) queryParams.append('pageSize', params.pageSize.toString());
      
      const query = queryParams.toString();
      return api.get<AuditListResponse>(`/api/laojun/audits${query ? `?${query}` : ''}`);
    },

    // 获取审计日志详情
    get: async (id: number): Promise<AuditLog> => {
      return api.get<AuditLog>(`/api/laojun/audits/${id}`);
    },
  },

  // 配置管理
  configs: {
    // 获取配置列表
    list: async (params?: {
      key?: string;
      page?: number;
      pageSize?: number;
    }): Promise<ConfigListResponse> => {
      const queryParams = new URLSearchParams();
      if (params?.key) queryParams.append('key', params.key);
      if (params?.page) queryParams.append('page', params.page.toString());
      if (params?.pageSize) queryParams.append('pageSize', params.pageSize.toString());
      
      const query = queryParams.toString();
      return api.get<ConfigListResponse>(`/api/laojun/configs${query ? `?${query}` : ''}`);
    },

    // 获取配置详情
    get: async (key: string): Promise<Config> => {
      return api.get<Config>(`/api/laojun/configs/${key}`);
    },

    // 更新配置
    update: async (key: string, data: ConfigUpdateRequest): Promise<{ updated: boolean; key: string }> => {
      return api.put<{ updated: boolean; key: string }>(`/api/laojun/configs/${key}`, data);
    },
  },

  // 模型管理
  models: {
    // 创建模型
    create: async (data: ModelCreateRequest): Promise<Model> => {
      return api.post<Model>('/api/taishang/models', data);
    },

    // 获取模型列表
    list: async (params?: {
      status?: string;
      name?: string;
      page?: number;
      pageSize?: number;
    }): Promise<ModelListResponse> => {
      const queryParams = new URLSearchParams();
      if (params?.status) queryParams.append('status', params.status);
      if (params?.name) queryParams.append('name', params.name);
      if (params?.page) queryParams.append('page', params.page.toString());
      if (params?.pageSize) queryParams.append('pageSize', params.pageSize.toString());
      
      const query = queryParams.toString();
      return api.get<ModelListResponse>(`/api/taishang/models${query ? `?${query}` : ''}`);
    },

    // 获取模型详情
    get: async (id: string): Promise<Model> => {
      return api.get<Model>(`/api/taishang/models/${id}`);
    },

    // 启用模型
    enable: async (id: string): Promise<{ id: string; status: string }> => {
      return api.post<{ id: string; status: string }>(`/api/taishang/models/${id}/enable`);
    },

    // 禁用模型
    disable: async (id: string): Promise<{ id: string; status: string }> => {
      return api.post<{ id: string; status: string }>(`/api/taishang/models/${id}/disable`);
    },
  },

  // 向量集合管理
  vectors: {
    // 创建向量集合
    createCollection: async (data: VectorCollectionCreateRequest): Promise<VectorCollection> => {
      return api.post<VectorCollection>('/api/taishang/vectors/collections', data);
    },

    // 获取向量集合列表
    listCollections: async (params?: {
      name?: string;
      page?: number;
      pageSize?: number;
    }): Promise<VectorCollectionListResponse> => {
      const queryParams = new URLSearchParams();
      if (params?.name) queryParams.append('name', params.name);
      if (params?.page) queryParams.append('page', params.page.toString());
      if (params?.pageSize) queryParams.append('pageSize', params.pageSize.toString());
      
      const query = queryParams.toString();
      return api.get<VectorCollectionListResponse>(`/api/taishang/vectors/collections${query ? `?${query}` : ''}`);
    },

    // 获取向量集合详情
    getCollection: async (id: string): Promise<VectorCollection> => {
      return api.get<VectorCollection>(`/api/taishang/vectors/collections/${id}`);
    },

    // 向集合中插入向量
    upsert: async (id: string, data: VectorUpsertRequest): Promise<VectorUpsertResponse> => {
      return api.post<VectorUpsertResponse>(`/api/taishang/vectors/collections/${id}/upsert`, data);
    },

    // 查询向量
    query: async (id: string, data: VectorQueryRequest): Promise<VectorQueryResponse> => {
      return api.post<VectorQueryResponse>(`/api/taishang/vectors/collections/${id}/query`, data);
    },

    // 删除向量
    delete: async (id: string, data: { ids: string[]; namespace?: string }): Promise<{ deleted: number }> => {
      return api.post<{ deleted: number }>(`/api/taishang/vectors/collections/${id}/delete`, data);
    },
  },

  // 任务管理
  tasks: {
    // 创建任务
    create: async (data: TaskCreateRequest): Promise<Task> => {
      return api.post<Task>('/api/taishang/tasks', data);
    },

    // 获取任务列表
    list: async (params?: {
      status?: string;
      type?: string;
      page?: number;
      pageSize?: number;
    }): Promise<TaskListResponse> => {
      const queryParams = new URLSearchParams();
      if (params?.status) queryParams.append('status', params.status);
      if (params?.type) queryParams.append('type', params.type);
      if (params?.page) queryParams.append('page', params.page.toString());
      if (params?.pageSize) queryParams.append('pageSize', params.pageSize.toString());
      
      const query = queryParams.toString();
      return api.get<TaskListResponse>(`/api/taishang/tasks${query ? `?${query}` : ''}`);
    },

    // 获取任务详情
    get: async (id: string): Promise<Task> => {
      return api.get<Task>(`/api/taishang/tasks/${id}`);
    },

    // 取消任务
    cancel: async (id: string): Promise<{ id: string; status: string }> => {
      return api.post<{ id: string; status: string }>(`/api/taishang/tasks/${id}/cancel`);
    },
  },
};