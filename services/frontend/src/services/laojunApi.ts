import { api } from './api';
import type { ApiResponse } from './api';

// 配置项接口
export interface Config {
  key: string;
  value: string;
  description?: string;
}

// 插件状态枚举
export enum PluginStatus {
  Installed = 'installed',
  Running = 'running',
  Stopped = 'stopped',
  Error = 'error'
}

// 插件接口
export interface Plugin {
  id: string;
  name: string;
  version: string;
  status: PluginStatus;
  description?: string;
  author?: string;
  homepage?: string;
  manifest?: any;
  installedAt?: string;
  updatedAt?: string;
}

// 审计日志接口
export interface AuditLog {
  id: string;
  tenantId: string;
  userId: string;
  action: string;
  resource: string;
  resourceId?: string;
  details?: any;
  ipAddress?: string;
  userAgent?: string;
  timestamp: string;
}

// 配置管理API
export const configApi = {
  // 获取所有配置
  getAll: (): Promise<ApiResponse<Config[]>> => {
    return api.get<Config[]>('/api/laojun/configs');
  },
  
  // 获取单个配置
  get: (key: string): Promise<ApiResponse<Config>> => {
    return api.get<Config>(`/api/laojun/configs/${key}`);
  },
  
  // 创建或更新配置
  set: (key: string, value: string, description?: string): Promise<ApiResponse<Config>> => {
    return api.post<Config>(`/api/laojun/configs/${key}`, { value, description });
  },
  
  // 删除配置
  delete: (key: string): Promise<ApiResponse<void>> => {
    return api.delete<void>(`/api/laojun/configs/${key}`);
  }
};

// 插件管理API
export const pluginApi = {
  // 获取所有插件
  getAll: (): Promise<ApiResponse<Plugin[]>> => {
    return api.get<Plugin[]>('/api/laojun/plugins/list');
  },
  
  // 获取单个插件
  get: (id: string): Promise<ApiResponse<Plugin>> => {
    return api.get<Plugin>(`/api/laojun/plugins/${id}`);
  },
  
  // 安装插件
  install: (id: string, source: string, version?: string): Promise<ApiResponse<Plugin>> => {
    return api.post<Plugin>(`/api/laojun/plugins/install`, { pluginId: id, source, version });
  },
  
  // 启动插件
  start: (id: string): Promise<ApiResponse<Plugin>> => {
    return api.post<Plugin>(`/api/laojun/plugins/start`, { pluginId: id });
  },
  
  // 停止插件
  stop: (id: string): Promise<ApiResponse<Plugin>> => {
    return api.post<Plugin>(`/api/laojun/plugins/stop`, { pluginId: id });
  },
  
  // 升级插件
  upgrade: (id: string, version?: string): Promise<ApiResponse<Plugin>> => {
    return api.post<Plugin>(`/api/laojun/plugins/upgrade`, { pluginId: id, version });
  },
  
  // 卸载插件
  uninstall: (id: string): Promise<ApiResponse<void>> => {
    return api.delete<void>(`/api/laojun/plugins/uninstall`, { data: { pluginId: id } });
  }
};

// 审计日志API
export const auditLogApi = {
  // 获取审计日志列表
  getAll: (page?: number, pageSize?: number): Promise<ApiResponse<{ logs: AuditLog[], total: number }>> => {
    const params = new URLSearchParams();
    if (page) params.append('page', page.toString());
    if (pageSize) params.append('pageSize', pageSize.toString());
    
    const queryString = params.toString();
    const url = queryString ? `/api/laojun/audits?${queryString}` : '/api/laojun/audits';
    
    return api.get<{ logs: AuditLog[], total: number }>(url);
  },
  
  // 获取单个审计日志
  get: (id: string): Promise<ApiResponse<AuditLog>> => {
    return api.get<AuditLog>(`/api/laojun/audits/${id}`);
  }
};