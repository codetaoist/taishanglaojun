import { apiClient } from './api';

export interface APICategory {
  id: number;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
}

export interface APIEndpoint {
  id: number;
  name: string;
  method: string;
  path: string;
  description: string;
  request_example?: string;
  response_example?: string;
  category_id: number;
  category?: APICategory;
  source_file: string;
  created_at: string;
  updated_at: string;
}

export interface CategoryListResponse {
  categories: APICategory[];
  total: number;
}

export interface EndpointListResponse {
  endpoints: APIEndpoint[];
  total: number;
}

export interface APIStatistics {
  total_categories: number;
  total_endpoints: number;
  endpoints_by_method: Record<string, number>;
  endpoints_by_category: Record<string, number>;
}

export interface APITestRequest {
  endpoint_id: number;
  test_type: string;
  request_data?: string;
  environment?: string;
}

export interface APITestRecord {
  id: number;
  endpoint_id: number;
  test_type: string;
  request_data?: string;
  response_data?: string;
  status_code?: number;
  response_time?: number;
  is_success: boolean;
  error_msg?: string;
  environment?: string;
  created_at: string;
  created_by: number;
}

class APIDocumentationService {
  private baseUrl = '/api/documentation';

  // 获取分类列表
  async getCategories(page = 1, pageSize = 50): Promise<CategoryListResponse> {
    const response = await apiClient.get(`${this.baseUrl}/categories`, {
      params: { page, page_size: pageSize }
    });
    return response.data;
  }

  // 根据ID获取分类详情
  async getCategoryById(id: number): Promise<APICategory> {
    const response = await apiClient.get(`${this.baseUrl}/categories/${id}`);
    return response.data;
  }

  // 获取接口列表
  async getEndpoints(page = 1, pageSize = 50): Promise<EndpointListResponse> {
    const response = await apiClient.get(`${this.baseUrl}/endpoints`, {
      params: { page, page_size: pageSize }
    });
    return response.data;
  }

  // 根据ID获取接口详情
  async getEndpointById(id: number): Promise<APIEndpoint> {
    const response = await apiClient.get(`${this.baseUrl}/endpoints/${id}`);
    return response.data;
  }

  // 根据分类获取接口列表
  async getEndpointsByCategory(categoryId: number, page = 1, pageSize = 50): Promise<EndpointListResponse> {
    const response = await apiClient.get(`${this.baseUrl}/categories/${categoryId}/endpoints`, {
      params: { page, page_size: pageSize }
    });
    return response.data;
  }

  // 搜索接口
  async searchEndpoints(query: string, page = 1, pageSize = 50): Promise<EndpointListResponse> {
    const response = await apiClient.get(`${this.baseUrl}/endpoints/search`, {
      params: { q: query, page, page_size: pageSize }
    });
    return response.data;
  }

  // 获取统计信息
  async getStatistics(): Promise<APIStatistics> {
    const response = await apiClient.get(`${this.baseUrl}/statistics`);
    return response.data;
  }

  // 测试API接口
  async testAPI(request: APITestRequest): Promise<APITestRecord> {
    const response = await apiClient.post(`${this.baseUrl}/test`, request);
    return response.data;
  }

  // 获取API测试历史
  async getTestHistory(endpointId: number, page = 1, pageSize = 20): Promise<{ records: APITestRecord[], total: number }> {
    const response = await apiClient.get(`${this.baseUrl}/test/history/${endpointId}`, {
      params: { page, page_size: pageSize }
    });
    return response.data;
  }
}

export const apiDocumentationService = new APIDocumentationService();
export default apiDocumentationService;