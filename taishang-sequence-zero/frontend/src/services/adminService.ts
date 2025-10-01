import { apiClient } from './authService';
import { User } from '@/types/auth';

// 系统统计接口
export interface SystemStats {
  totalUsers: number;
  activeUsers: number;
  totalSessions: number;
  systemLoad: number;
  memoryUsage: number;
  diskUsage: number;
}

// 用户管理接口
export interface UserManagement {
  users: User[];
  totalCount: number;
  currentPage: number;
  pageSize: number;
}

// 系统日志接口
export interface SystemLog {
  id: string;
  timestamp: string;
  level: 'info' | 'warn' | 'error' | 'debug';
  message: string;
  source: string;
  userId?: string;
}

// 系统配置接口
export interface SystemConfig {
  siteName: string;
  siteDescription: string;
  maxUsers: number;
  sessionTimeout: number;
  enableRegistration: boolean;
  enableGuestAccess: boolean;
}

class AdminService {
  // 获取系统统计信息
  async getSystemStats(): Promise<SystemStats> {
    const response = await apiClient.get('/admin/stats');
    return response.data;
  }

  // 获取用户列表
  async getUsers(params: { page?: number; pageSize?: number; search?: string; status?: string } = {}): Promise<UserManagement> {
    const queryParams = {
      page: params.page || 1,
      pageSize: params.pageSize || 10,
      search: params.search,
      status: params.status
    };
    const response = await apiClient.get('/admin/users', { params: queryParams });
    return response.data;
  }

  // 创建用户
  async createUser(userData: Partial<User>): Promise<User> {
    const response = await apiClient.post('/admin/users', userData);
    return response.data;
  }

  // 更新用户
  async updateUser(userId: string, userData: Partial<User>): Promise<User> {
    const response = await apiClient.put(`/admin/users/${userId}`, userData);
    return response.data;
  }

  // 删除用户
  async deleteUser(userId: string): Promise<void> {
    await apiClient.delete(`/admin/users/${userId}`);
  }

  // 批量删除用户
  async deleteUsers(userIds: string[]): Promise<void> {
    await apiClient.post('/admin/users/batch-delete', { userIds });
  }

  // 重置用户密码
  async resetUserPassword(userId: string): Promise<{ tempPassword: string }> {
    const response = await apiClient.post(`/admin/users/${userId}/reset-password`);
    return response.data;
  }

  // 更新用户状态
  async updateUserStatus(params: { userId: string; status: string; reason?: string }): Promise<User> {
    const response = await apiClient.put(`/admin/users/${params.userId}/status`, {
      status: params.status,
      reason: params.reason
    });
    return response.data;
  }

  // 获取系统日志
  async getSystemLogs(params: {
    page?: number;
    pageSize?: number;
    level?: string;
    source?: string;
    limit?: number;
    offset?: number;
    startDate?: string;
    endDate?: string;
  } = {}): Promise<{ logs: SystemLog[]; totalCount: number }> {
    const queryParams = {
      page: params.page || 1,
      pageSize: params.pageSize || params.limit || 50,
      level: params.level,
      source: params.source,
      offset: params.offset,
      startDate: params.startDate,
      endDate: params.endDate
    };
    const response = await apiClient.get('/admin/logs', { params: queryParams });
    return response.data;
  }

  // 清理系统日志
  async clearSystemLogs(beforeDate?: string): Promise<void> {
    await apiClient.delete('/admin/logs', { data: { beforeDate } });
  }

  // 获取系统配置
  async getSystemConfig(): Promise<SystemConfig> {
    const response = await apiClient.get('/admin/config');
    return response.data;
  }

  // 更新系统配置
  async updateSystemConfig(config: Partial<SystemConfig>): Promise<SystemConfig> {
    const response = await apiClient.put('/admin/config', config);
    return response.data;
  }

  // 系统备份
  async createBackup(): Promise<{ backupId: string; filename: string }> {
    const response = await apiClient.post('/admin/backup');
    return response.data;
  }

  // 获取备份列表
  async getBackups(): Promise<Array<{ id: string; filename: string; createdAt: string; size: number }>> {
    const response = await apiClient.get('/admin/backups');
    return response.data;
  }

  // 恢复备份
  async restoreBackup(backupId: string): Promise<void> {
    await apiClient.post(`/admin/backups/${backupId}/restore`);
  }

  // 删除备份
  async deleteBackup(backupId: string): Promise<void> {
    await apiClient.delete(`/admin/backups/${backupId}`);
  }

  // 系统健康检查
  async healthCheck(): Promise<{
    status: 'healthy' | 'warning' | 'critical';
    checks: Array<{
      name: string;
      status: 'pass' | 'fail' | 'warn';
      message?: string;
    }>;
  }> {
    const response = await apiClient.get('/admin/health');
    return response.data;
  }

  // 获取角色列表
  async getRoles(): Promise<Array<{
    id: string;
    name: string;
    description: string;
    permissions: string[];
    level: number;
    color: string;
  }>> {
    const response = await apiClient.get('/admin/roles');
    return response.data;
  }

  // 获取权限列表
  async getPermissions(): Promise<Array<{
    id: string;
    name: string;
    description: string;
    category: string;
    level: number;
  }>> {
    const response = await apiClient.get('/admin/permissions');
    return response.data;
  }
}

const adminService = new AdminService();
export default adminService;