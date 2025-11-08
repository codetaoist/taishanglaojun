import { api } from './api';
import type { ApiResponse } from './api';
import type { Collection } from './taishangApi';

// 向量数据接口
export interface VectorData {
  id: string;
  vector: number[];
  metadata?: Record<string, any>;
  createdAt?: string;
  updatedAt?: string;
}

// 向量数据列表响应接口
export interface VectorDataListResponse {
  total: number;
  page: number;
  pageSize: number;
  items: VectorData[];
}

// 向量搜索请求接口
export interface VectorSearchRequest {
  collectionName: string;
  vector: number[];
  topK: number;
  filter?: Record<string, any>;
}

// 向量搜索结果接口
export interface VectorSearchResult {
  id: string;
  score: number;
  vector?: number[];
  metadata?: Record<string, any>;
}

// 向量搜索响应接口
export interface VectorSearchResponse {
  results: VectorSearchResult[];
  total: number;
}

// 向量数据管理API
export const vectorDataApi = {
  // 获取集合中的向量数据列表
  getAll: (collectionName: string, page = 1, pageSize = 10): Promise<ApiResponse<VectorDataListResponse>> => {
    return api.get<VectorDataListResponse>(`/api/v1/vector/collections/${collectionName}/vectors?page=${page}&pageSize=${pageSize}`);
  },
  
  // 获取单个向量数据
  get: (collectionName: string, vectorId: string): Promise<ApiResponse<VectorData>> => {
    return api.get<VectorData>(`/api/v1/vector/collections/${collectionName}/vectors/${vectorId}`);
  },
  
  // 插入或更新向量数据
  upsert: (collectionName: string, vectors: VectorData[]): Promise<ApiResponse<{ upserted: number }>> => {
    return api.post<{ upserted: number }>(`/api/v1/vector/collections/${collectionName}/vectors`, { vectors });
  },
  
  // 删除向量数据
  delete: (collectionName: string, vectorIds?: string[]): Promise<ApiResponse<{ deleted: number }>> => {
    const payload = vectorIds ? { ids: vectorIds } : {};
    return api.delete<{ deleted: number }>(`/api/v1/vector/collections/${collectionName}/vectors`, { data: payload });
  },
  
  // 搜索向量数据
  search: (request: VectorSearchRequest): Promise<ApiResponse<VectorSearchResponse>> => {
    return api.post<VectorSearchResponse>(`/api/v1/vector/collections/${request.collectionName}/search`, request);
  }
};