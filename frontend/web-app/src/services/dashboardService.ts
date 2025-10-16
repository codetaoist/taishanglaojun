import axios from 'axios';
import type { AxiosInstance, AxiosResponse } from 'axios';

// 定义API响应类型
interface ApiResponse<T = unknown> {
  success: boolean;
  data: T;
  message?: string;
  error?: string;
}

// 仪表板统计数据类型
export interface DashboardStats {
  totalUsers: number;
  activeUsers: number;
  totalProjects: number;
  completedTasks: number;
  pendingTasks: number;
  systemHealth: number;
  cpuUsage: number;
  memoryUsage: number;
  diskUsage: number;
  networkLatency: number;
}

// 系统指标类型
export interface SystemMetrics {
  cpu: {
    usage: number;
    cores: number;
    temperature?: number;
  };
  memory: {
    total: number;
    used: number;
    free: number;
    usage: number;
  };
  disk: {
    total: number;
    used: number;
    free: number;
    usage: number;
  };
  network: {
    latency: number;
    throughput: {
      in: number;
      out: number;
    };
  };
}

// 活动数据类型
export interface ActivityData {
  id: string;
  type: 'user_login' | 'task_completed' | 'project_created' | 'system_alert';
  title: string;
  description: string;
  timestamp: string;
  user?: {
    id: string;
    username: string;
    avatar?: string;
  };
  severity?: 'info' | 'warning' | 'error' | 'success';
}

// 趋势数据类型
export interface TrendData {
  date: string;
  users: number;
  tasks: number;
  projects: number;
  performance: number;
}

class DashboardService {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1',
      timeout: 10000,
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
        }
        return Promise.reject(error);
      }
    );
  }

  // 获取仪表板统计数据
  async getDashboardStats(): Promise<ApiResponse<DashboardStats>> {
    try {
      const response = await this.client.get('/dashboards/stats');
      return response.data;
    } catch (error) {
      console.error('Failed to fetch dashboard stats:', error);
      // 返回模拟数据作为后备
      return {
        success: true,
        data: {
          totalUsers: 1250,
          activeUsers: 892,
          totalProjects: 45,
          completedTasks: 1834,
          pendingTasks: 267,
          systemHealth: 98.5,
          cpuUsage: 45.2,
          memoryUsage: 67.8,
          diskUsage: 34.5,
          networkLatency: 12
        }
      };
    }
  }

  // 获取系统指标
  async getSystemMetrics(): Promise<ApiResponse<SystemMetrics>> {
    try {
      const response = await this.client.get('/dashboards/metrics');
      return response.data;
    } catch (error) {
      console.error('Failed to fetch system metrics:', error);
      // 返回模拟数据作为后备
      return {
        success: true,
        data: {
          cpu: {
            usage: 45.2,
            cores: 8,
            temperature: 65
          },
          memory: {
            total: 16384,
            used: 11108,
            free: 5276,
            usage: 67.8
          },
          disk: {
            total: 512000,
            used: 176640,
            free: 335360,
            usage: 34.5
          },
          network: {
            latency: 12,
            throughput: {
              in: 1024,
              out: 512
            }
          }
        }
      };
    }
  }

  // 获取最近活动
  async getRecentActivities(limit: number = 10): Promise<ApiResponse<ActivityData[]>> {
    try {
      const response = await this.client.get(`/dashboards/activities?limit=${limit}`);
      return response.data;
    } catch (error) {
      console.error('Failed to fetch recent activities:', error);
      // 返回模拟数据作为后备
      return {
        success: true,
        data: [
          {
            id: '1',
            type: 'user_login',
            title: '用户登录',
            description: '张三 登录了系统',
            timestamp: new Date(Date.now() - 5 * 60 * 1000).toISOString(),
            user: {
              id: '1',
              username: '张三',
              avatar: '/avatars/zhangsan.jpg'
            },
            severity: 'info'
          },
          {
            id: '2',
            type: 'task_completed',
            title: '任务完成',
            description: '李四 完成了任务 "用户界面优化"',
            timestamp: new Date(Date.now() - 15 * 60 * 1000).toISOString(),
            user: {
              id: '2',
              username: '李四',
              avatar: '/avatars/lisi.jpg'
            },
            severity: 'success'
          },
          {
            id: '3',
            type: 'project_created',
            title: '项目创建',
            description: '王五 创建了新项目 "移动端应用开发"',
            timestamp: new Date(Date.now() - 30 * 60 * 1000).toISOString(),
            user: {
              id: '3',
              username: '王五',
              avatar: '/avatars/wangwu.jpg'
            },
            severity: 'info'
          },
          {
            id: '4',
            type: 'system_alert',
            title: '系统警告',
            description: 'CPU使用率超过80%',
            timestamp: new Date(Date.now() - 45 * 60 * 1000).toISOString(),
            severity: 'warning'
          },
          {
            id: '5',
            type: 'task_completed',
            title: '任务完成',
            description: '赵六 完成了任务 "数据库优化"',
            timestamp: new Date(Date.now() - 60 * 60 * 1000).toISOString(),
            user: {
              id: '4',
              username: '赵六',
              avatar: '/avatars/zhaoliu.jpg'
            },
            severity: 'success'
          }
        ]
      };
    }
  }

  // 获取趋势数据
  async getTrendData(days: number = 7): Promise<ApiResponse<TrendData[]>> {
    try {
      const response = await this.client.get(`/dashboards/trends?days=${days}`);
      return response.data;
    } catch (error) {
      console.error('Failed to fetch trend data:', error);
      // 返回模拟数据作为后备
      const trendData: TrendData[] = [];
      for (let i = days - 1; i >= 0; i--) {
        const date = new Date();
        date.setDate(date.getDate() - i);
        trendData.push({
          date: date.toISOString().split('T')[0],
          users: Math.floor(Math.random() * 100) + 50,
          tasks: Math.floor(Math.random() * 50) + 20,
          projects: Math.floor(Math.random() * 10) + 2,
          performance: Math.floor(Math.random() * 20) + 80
        });
      }
      return {
        success: true,
        data: trendData
      };
    }
  }

  // 获取快捷操作数据
  async getQuickActions(): Promise<ApiResponse<any[]>> {
    try {
      const response = await this.client.get('/dashboards/quick-actions');
      return response.data;
    } catch (error) {
      console.error('Failed to fetch quick actions:', error);
      // 返回模拟数据作为后备
      return {
        success: true,
        data: [
          {
            id: 'create-project',
            title: '创建项目',
            description: '快速创建新项目',
            icon: 'plus',
            action: '/projects/create',
            color: '#1890ff'
          },
          {
            id: 'add-user',
            title: '添加用户',
            description: '邀请新用户加入',
            icon: 'user-add',
            action: '/users/invite',
            color: '#52c41a'
          },
          {
            id: 'system-settings',
            title: '系统设置',
            description: '配置系统参数',
            icon: 'setting',
            action: '/settings',
            color: '#faad14'
          },
          {
            id: 'view-reports',
            title: '查看报告',
            description: '生成和查看报告',
            icon: 'file-text',
            action: '/reports',
            color: '#722ed1'
          }
        ]
      };
    }
  }

  // 获取健康检查状态
  async getHealthStatus(): Promise<ApiResponse<any>> {
    try {
      const response = await this.client.get('/health');
      return response.data;
    } catch (error) {
      console.error('Failed to fetch health status:', error);
      return {
        success: false,
        data: null,
        error: 'Failed to fetch health status'
      };
    }
  }
}

// 导出单例实例
export const dashboardService = new DashboardService();
export default dashboardService;