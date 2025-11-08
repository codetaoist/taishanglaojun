import { api } from './api';

export interface VectorDatabaseStatus {
  connected: boolean;
  lastChecked: string;
  error?: string;
}

export interface VectorDatabaseInfo {
  type: string;
  version: string;
}

export interface CollectionStats {
  count: number;
}

export interface VectorCollectionInfo {
  name: string;
  description?: string;
  dimension?: number;
  metricType?: string;
  vectorCount?: number;
}

export const vectorMonitorApi = {
  // 获取向量数据库状态
  getStatus: async (): Promise<VectorDatabaseStatus> => {
    const response = await api.get('/vector/status');
    return response.data;
  },

  // 获取向量数据库信息
  getInfo: async (): Promise<VectorDatabaseInfo> => {
    const response = await api.get('/vector/info');
    return response.data;
  },

  // 健康检查
  healthCheck: async (): Promise<{ status: string }> => {
    const response = await api.get('/vector/health');
    return response.data;
  },

  // 连接向量数据库
  connect: async (): Promise<{ message: string }> => {
    const response = await api.post('/vector/connect');
    return response.data;
  },

  // 获取集合统计信息
  getCollectionStats: async (collectionName: string): Promise<CollectionStats> => {
    const response = await api.get(`/vector/collections/${collectionName}/stats`);
    return response.data;
  },

  // 列出所有向量集合
  listVectorCollections: async (): Promise<VectorCollectionInfo[]> => {
    const response = await api.get('/vector/collections');
    return response.data;
  },

  // 获取特定集合信息
  getVectorCollection: async (collectionName: string): Promise<VectorCollectionInfo> => {
    const response = await api.get(`/vector/collections/${collectionName}`);
    return response.data;
  },
};