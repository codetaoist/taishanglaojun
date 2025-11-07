import { get, post, put, del } from './api';
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
    return get<Config[]>('/laojun/config');
  },
  
  // 获取单个配置
  get: (key: string): Promise<ApiResponse<Config>> => {
    return get<Config>(`/laojun/config/${key}`);
  },
  
  // 创建或更新配置
  set: (key: string, value: string, description?: string): Promise<ApiResponse<Config>> => {
    return post<Config>(`/laojun/config/${key}`, { value, description });
  },
  
  // 删除配置
  delete: (key: string): Promise<ApiResponse<void>> => {
    return del<void>(`/laojun/config/${key}`);
  }
};

// 插件管理API
export const pluginApi = {
  // 获取所有插件
  getAll: (): Promise<ApiResponse<Plugin[]>> => {
    return get<Plugin[]>('/laojun/plugins');
  },
  
  // 获取单个插件
  get: (id: string): Promise<ApiResponse<Plugin>> => {
    return get<Plugin>(`/laojun/plugins/${id}`);
  },
  
  // 安装插件
  install: (id: string, source: string, version?: string): Promise<ApiResponse<Plugin>> => {
    return post<Plugin>(`/laojun/plugins/${id}/install`, { source, version });
  },
  
  // 启动插件
  start: (id: string): Promise<ApiResponse<Plugin>> => {
    return post<Plugin>(`/laojun/plugins/${id}/start`);
  },
  
  // 停止插件
  stop: (id: string): Promise<ApiResponse<Plugin>> => {
    return post<Plugin>(`/laojun/plugins/${id}/stop`);
  },
  
  // 升级插件
  upgrade: (id: string, version?: string): Promise<ApiResponse<Plugin>> => {
    return post<Plugin>(`/laojun/plugins/${id}/upgrade`, { version });
  },
  
  // 卸载插件
  uninstall: (id: string): Promise<ApiResponse<void>> => {
    return del<void>(`/laojun/plugins/${id}`);
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
    const url = queryString ? `/laojun/audit-logs?${queryString}` : '/laojun/audit-logs';
    
    return get<{ logs: AuditLog[], total: number }>(url);
  },
  
  // 获取单个审计日志
  get: (id: string): Promise<ApiResponse<AuditLog>> => {
    return get<AuditLog>(`/laojun/audit-logs/${id}`);
  }
};