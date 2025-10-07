/**
 * 太上老君AI平台 - 系统菜单配置
 * 基于系统菜单架构设计，集成所有功能模块
 */

import React from 'react';
import {
  HomeOutlined,
  MessageOutlined,
  BookOutlined,
  TeamOutlined,
  ReadOutlined,
  ProjectOutlined,
  HeartOutlined,
  SecurityScanOutlined,
  SettingOutlined,
  MobileOutlined,
  SearchOutlined,
  RobotOutlined,
  EditOutlined,
  UserOutlined,
  DashboardOutlined,
  BulbOutlined,
  BarChartOutlined,
  FileTextOutlined,
  CameraOutlined,
  SoundOutlined,
  VideoCameraOutlined,
  TagOutlined,
  StarOutlined,
  CommentOutlined,
  CalendarOutlined,
  FolderOutlined,
  CheckSquareOutlined,
  UsergroupAddOutlined,
  LineChartOutlined,
  MonitorOutlined,
  ExperimentOutlined,
  TrophyOutlined,
  SafetyCertificateOutlined,
  EyeOutlined,
  ToolOutlined,
  CloudOutlined,
  DesktopOutlined,
  TabletOutlined,
  GlobalOutlined,
  ApiOutlined,
  BellOutlined,
  LockOutlined,
  DatabaseOutlined,
  ThunderboltOutlined,
  FireOutlined,
  RocketOutlined,
  AuditOutlined
} from '@ant-design/icons';

export interface MenuItem {
  key: string;
  icon?: React.ReactNode;
  label: string;
  path?: string;
  children?: MenuItem[];
  status: 'completed' | 'partial' | 'planned';
  description?: string;
  requiredRole?: string[];
  requiredPermission?: string[];
  badge?: string;
  priority: 'high' | 'medium' | 'low';
}

/**
 * 主菜单配置
 * 状态说明：
 * - completed: ✅ 已完成开发
 * - partial: 🔄 部分开发
 * - planned: ⏳ 规划中
 */
export const mainMenuConfig: MenuItem[] = [
  // 1. 仪表板
  {
    key: 'dashboard',
    icon: <DashboardOutlined />,
    label: '仪表板',
    path: '/',
    status: 'completed',
    description: '系统概览、快捷操作、个性化推荐',
    priority: 'high'
  },

  // 2. AI智能服务
  {
    key: 'ai-services',
    icon: <RobotOutlined />,
    label: 'AI智能服务',
    status: 'partial',
    description: 'AI对话、多模态AI、智能分析、内容生成',
    priority: 'high',
    children: [
      {
        key: 'ai-chat',
        icon: <MessageOutlined />,
        label: '智能对话',
        path: '/chat',
        status: 'completed',
        description: '多轮对话、专业领域、语音交互',
        priority: 'high'
      },
      {
        key: 'ai-multimodal',
        icon: <CameraOutlined />,
        label: '多模态AI',
        path: '/ai/multimodal',
        status: 'partial',
        description: '图像生成、图像分析、视频处理、音频处理',
        priority: 'high',
        children: [
          {
            key: 'image-generation',
            icon: <CameraOutlined />,
            label: '图像生成',
            path: '/ai/multimodal/image-generation',
            status: 'planned',
            priority: 'medium'
          },
          {
            key: 'image-analysis',
            icon: <EyeOutlined />,
            label: '图像分析',
            path: '/ai/multimodal/image-analysis',
            status: 'planned',
            priority: 'medium'
          },
          {
            key: 'video-processing',
            icon: <VideoCameraOutlined />,
            label: '视频处理',
            path: '/ai/multimodal/video-processing',
            status: 'planned',
            priority: 'low'
          },
          {
            key: 'audio-processing',
            icon: <SoundOutlined />,
            label: '音频处理',
            path: '/ai/multimodal/audio-processing',
            status: 'planned',
            priority: 'low'
          }
        ]
      },
      {
        key: 'ai-analysis',
        icon: <BarChartOutlined />,
        label: '智能分析',
        path: '/ai/analysis',
        status: 'planned',
        description: '数据分析、趋势预测、报告生成',
        priority: 'medium',
        children: [
          {
            key: 'data-analysis',
            icon: <LineChartOutlined />,
            label: '数据分析',
            path: '/ai/analysis/data',
            status: 'planned',
            priority: 'medium'
          },
          {
            key: 'trend-prediction',
            icon: <ThunderboltOutlined />,
            label: '趋势预测',
            path: '/ai/analysis/trends',
            status: 'planned',
            priority: 'medium'
          },
          {
            key: 'report-generation',
            icon: <FileTextOutlined />,
            label: '报告生成',
            path: '/ai/analysis/reports',
            status: 'planned',
            priority: 'low'
          }
        ]
      },
      {
        key: 'ai-generation',
        icon: <BulbOutlined />,
        label: '内容生成',
        path: '/ai/generation',
        status: 'planned',
        description: '文本生成、代码生成、创意设计',
        priority: 'medium',
        children: [
          {
            key: 'text-generation',
            icon: <EditOutlined />,
            label: '文本生成',
            path: '/ai/generation/text',
            status: 'planned',
            priority: 'medium'
          },
          {
            key: 'code-generation',
            icon: <ApiOutlined />,
            label: '代码生成',
            path: '/ai/generation/code',
            status: 'planned',
            priority: 'medium'
          },
          {
            key: 'creative-design',
            icon: <FireOutlined />,
            label: '创意设计',
            path: '/ai/generation/design',
            status: 'planned',
            priority: 'low'
          }
        ]
      }
    ]
  },

  // 3. 文化智慧
  {
    key: 'cultural-wisdom',
    icon: <BookOutlined />,
    label: '文化智慧',
    status: 'completed',
    description: '智慧库、智慧搜索、智慧推荐、分类管理',
    priority: 'high',
    children: [
      {
        key: 'wisdom-library',
        icon: <BookOutlined />,
        label: '智慧库',
        path: '/wisdom',
        status: 'completed',
        description: '经典文献、现代解读、智慧问答',
        priority: 'high'
      },
      {
        key: 'wisdom-search',
        icon: <SearchOutlined />,
        label: '智慧搜索',
        path: '/wisdom/search',
        status: 'completed',
        description: '语义搜索、关联推荐、搜索历史',
        priority: 'high'
      },
      {
        key: 'wisdom-recommend',
        icon: <StarOutlined />,
        label: '智慧推荐',
        path: '/recommendations',
        status: 'completed',
        description: '个性化推荐、每日智慧、主题推荐',
        priority: 'high'
      },
      {
        key: 'wisdom-category',
        icon: <TagOutlined />,
        label: '分类管理',
        path: '/wisdom/categories',
        status: 'completed',
        description: '分类体系、标签管理、内容管理',
        priority: 'medium',
        requiredRole: ['admin', 'editor']
      },
      {
        key: 'wisdom-favorites',
        icon: <HeartOutlined />,
        label: '我的收藏',
        path: '/favorites',
        status: 'completed',
        description: '收藏管理、个人智慧库',
        priority: 'medium'
      },
      {
        key: 'wisdom-notes',
        icon: <EditOutlined />,
        label: '我的笔记',
        path: '/notes',
        status: 'completed',
        description: '学习笔记、心得体会',
        priority: 'medium'
      }
    ]
  },

  // 4. 社区交流
  {
    key: 'community',
    icon: <TeamOutlined />,
    label: '社区交流',
    status: 'completed',
    description: '社区动态、实时聊天、兴趣小组、活动中心',
    priority: 'high',
    children: [
      {
        key: 'community-posts',
        icon: <CommentOutlined />,
        label: '社区动态',
        path: '/community',
        status: 'completed',
        description: '动态发布、互动功能、话题讨论',
        priority: 'high'
      },
      {
        key: 'community-chat',
        icon: <MessageOutlined />,
        label: '实时聊天',
        path: '/community/chat',
        status: 'completed',
        description: '私聊功能、群聊功能、文件传输',
        priority: 'high'
      },
      {
        key: 'community-groups',
        icon: <UsergroupAddOutlined />,
        label: '兴趣小组',
        path: '/community/groups',
        status: 'completed',
        description: '小组创建、成员管理、活动组织',
        priority: 'medium'
      },
      {
        key: 'community-events',
        icon: <CalendarOutlined />,
        label: '活动中心',
        path: '/community/events',
        status: 'planned',
        description: '活动发布、报名管理、活动直播',
        priority: 'medium'
      }
    ]
  },

  // 5. 智能学习
  {
    key: 'intelligent-learning',
    icon: <ReadOutlined />,
    label: '智能学习',
    status: 'partial',
    description: '课程中心、学习进度、能力评估、认证中心',
    priority: 'high',
    children: [
      {
        key: 'learning-courses',
        icon: <BookOutlined />,
        label: '课程中心',
        path: '/learning/courses',
        status: 'partial',
        description: '课程目录、课程播放、课程笔记',
        priority: 'high'
      },
      {
        key: 'learning-progress',
        icon: <LineChartOutlined />,
        label: '学习进度',
        path: '/learning/progress',
        status: 'completed',
        description: '进度追踪、学习计划、学习统计',
        priority: 'high'
      },
      {
        key: 'learning-assessment',
        icon: <ExperimentOutlined />,
        label: '能力评估',
        path: '/learning/assessment',
        status: 'partial',
        description: '技能测试、能力分析、学习建议',
        priority: 'medium'
      },
      {
        key: 'learning-certificate',
        icon: <TrophyOutlined />,
        label: '认证中心',
        path: '/learning/certificates',
        status: 'planned',
        description: '证书管理、认证申请、证书验证',
        priority: 'medium'
      }
    ]
  },

  // 6. 项目管理
  {
    key: 'project-management',
    icon: <ProjectOutlined />,
    label: '项目管理',
    status: 'completed',
    description: '项目工作台、任务管理、团队协作、项目分析',
    priority: 'medium',
    children: [
      {
        key: 'project-workspace',
        icon: <FolderOutlined />,
        label: '项目工作台',
        path: '/projects/workspace',
        status: 'completed',
        description: '项目概览、项目创建、项目模板',
        priority: 'medium'
      },
      {
        key: 'project-tasks',
        icon: <CheckSquareOutlined />,
        label: '任务管理',
        path: '/projects/tasks',
        status: 'completed',
        description: '任务创建、任务跟踪、任务提醒',
        priority: 'high'
      },
      {
        key: 'project-collaboration',
        icon: <UsergroupAddOutlined />,
        label: '团队协作',
        path: '/projects/collaboration',
        status: 'completed',
        description: '团队管理、协作工具、文档共享',
        priority: 'medium'
      },
      {
        key: 'project-analytics',
        icon: <BarChartOutlined />,
        label: '项目分析',
        path: '/projects/analytics',
        status: 'completed',
        description: '项目报告、数据分析、风险评估',
        priority: 'low'
      }
    ]
  },

  // 7. 健康管理
  {
    key: 'health-management',
    icon: <HeartOutlined />,
    label: '健康管理',
    status: 'planned',
    description: '健康监测、健康分析、健康建议、健康档案',
    priority: 'high',
    children: [
      {
        key: 'health-monitor',
        icon: <MonitorOutlined />,
        label: '健康监测',
        path: '/health/monitor',
        status: 'planned',
        description: '生理监测、运动追踪、睡眠分析',
        priority: 'high'
      },
      {
        key: 'health-analysis',
        icon: <BarChartOutlined />,
        label: '健康分析',
        path: '/health/analysis',
        status: 'planned',
        description: '健康报告、趋势分析、异常预警',
        priority: 'high'
      },
      {
        key: 'health-advice',
        icon: <BulbOutlined />,
        label: '健康建议',
        path: '/health/advice',
        status: 'planned',
        description: '个性化建议、运动计划、饮食建议',
        priority: 'medium'
      },
      {
        key: 'health-records',
        icon: <FileTextOutlined />,
        label: '健康档案',
        path: '/health/records',
        status: 'planned',
        description: '档案管理、病历记录、体检报告',
        priority: 'medium'
      }
    ]
  },

  // 8. 安全中心
  {
    key: 'security-center',
    icon: <SecurityScanOutlined />,
    label: '安全中心',
    path: '/security',
    status: 'completed',
    description: '威胁检测、漏洞管理、渗透测试、安全教育、安全审计',
    priority: 'high',
    requiredRole: ['admin', 'security'],
    children: [
      {
        key: 'threat-detection',
        icon: <EyeOutlined />,
        label: '威胁检测',
        path: '/security#threat-detection',
        status: 'completed',
        description: '实时威胁监控、威胁告警、检测规则管理',
        priority: 'high',
        requiredRole: ['admin', 'security']
      },
      {
        key: 'vulnerability-management',
        icon: <SafetyCertificateOutlined />,
        label: '漏洞管理',
        path: '/security#vulnerability',
        status: 'completed',
        description: '漏洞扫描、漏洞评估、修复建议',
        priority: 'high',
        requiredRole: ['admin', 'security']
      },
      {
        key: 'penetration-testing',
        icon: <ToolOutlined />,
        label: '渗透测试',
        path: '/security#penetration-testing',
        status: 'completed',
        description: '渗透测试项目、测试结果、安全评估',
        priority: 'high',
        requiredRole: ['admin', 'security']
      },
      {
        key: 'security-education',
        icon: <ReadOutlined />,
        label: '安全教育',
        path: '/security#security-education',
        status: 'completed',
        description: '安全培训课程、实验环境、认证管理',
        priority: 'medium'
      },
      {
        key: 'security-audit',
        icon: <AuditOutlined />,
        label: '安全审计',
        path: '/security#security-audit',
        status: 'completed',
        description: '审计日志、安全事件、合规报告',
        priority: 'medium',
        requiredRole: ['admin', 'security']
      }
    ]
  },

  // 9. 系统管理
  {
    key: 'system-management',
    icon: <SettingOutlined />,
    label: '系统管理',
    status: 'completed',
    description: '系统设置、用户管理、权限管理、系统监控',
    priority: 'high',
    requiredRole: ['admin'],
    children: [
      {
        key: 'system-settings',
        icon: <SettingOutlined />,
        label: '系统设置',
        path: '/admin/settings',
        status: 'completed',
        description: '基础配置、功能开关、性能调优',
        priority: 'high',
        requiredRole: ['admin']
      },
      {
        key: 'system-users',
        icon: <UserOutlined />,
        label: '用户管理',
        path: '/admin/users',
        status: 'completed',
        description: '用户列表、用户权限、用户组',
        priority: 'high',
        requiredRole: ['admin']
      },
      {
        key: 'system-permissions',
        icon: <LockOutlined />,
        label: '权限管理',
        path: '/admin/permissions',
        status: 'completed',
        description: '角色管理、权限分配、访问控制',
        priority: 'high',
        requiredRole: ['admin']
      },
      {
        key: 'system-monitoring',
        icon: <MonitorOutlined />,
        label: '系统监控',
        path: '/admin/monitoring',
        status: 'completed',
        description: '性能监控、资源监控、日志管理',
        priority: 'high',
        requiredRole: ['admin']
      },
      {
        key: 'system-analytics',
        icon: <BarChartOutlined />,
        label: '数据分析',
        path: '/admin/analytics',
        status: 'completed',
        description: '用户分析、内容分析、系统分析',
        priority: 'medium',
        requiredRole: ['admin']
      },
      {
        key: 'system-notifications',
        icon: <BellOutlined />,
        label: '通知中心',
        path: '/admin/notifications',
        status: 'completed',
        description: '系统通知、消息推送、通知模板',
        priority: 'medium',
        requiredRole: ['admin']
      }
    ]
  },

  // 10. 跨平台应用
  {
    key: 'cross-platform',
    icon: <MobileOutlined />,
    label: '跨平台应用',
    status: 'partial',
    description: '移动应用、桌面应用、手表应用、Web应用',
    priority: 'medium',
    children: [
      {
        key: 'mobile-apps',
        icon: <MobileOutlined />,
        label: '移动应用',
        path: '/apps/mobile',
        status: 'partial',
        description: 'Android、iOS、鸿蒙应用',
        priority: 'high',
        children: [
          {
            key: 'android-app',
            icon: <MobileOutlined />,
            label: 'Android应用',
            path: '/apps/mobile/android',
            status: 'partial',
            priority: 'high'
          },
          {
            key: 'ios-app',
            icon: <MobileOutlined />,
            label: 'iOS应用',
            path: '/apps/mobile/ios',
            status: 'partial',
            priority: 'high'
          },
          {
            key: 'harmony-app',
            icon: <MobileOutlined />,
            label: '鸿蒙应用',
            path: '/apps/mobile/harmony',
            status: 'partial',
            priority: 'medium'
          }
        ]
      },
      {
        key: 'desktop-apps',
        icon: <DesktopOutlined />,
        label: '桌面应用',
        path: '/apps/desktop',
        status: 'partial',
        description: 'Windows、macOS、Linux应用',
        priority: 'medium',
        children: [
          {
            key: 'windows-app',
            icon: <DesktopOutlined />,
            label: 'Windows应用',
            path: '/apps/desktop/windows',
            status: 'planned',
            priority: 'medium'
          },
          {
            key: 'macos-app',
            icon: <DesktopOutlined />,
            label: 'macOS应用',
            path: '/apps/desktop/macos',
            status: 'partial',
            priority: 'medium'
          },
          {
            key: 'linux-app',
            icon: <DesktopOutlined />,
            label: 'Linux应用',
            path: '/apps/desktop/linux',
            status: 'partial',
            priority: 'medium'
          }
        ]
      },
      {
        key: 'watch-apps',
        icon: <TabletOutlined />,
        label: '手表应用',
        path: '/apps/watch',
        status: 'planned',
        description: 'Apple Watch、Wear OS、华为手表',
        priority: 'low',
        children: [
          {
            key: 'apple-watch',
            icon: <TabletOutlined />,
            label: 'Apple Watch',
            path: '/apps/watch/apple',
            status: 'planned',
            priority: 'low'
          },
          {
            key: 'wear-os',
            icon: <TabletOutlined />,
            label: 'Wear OS',
            path: '/apps/watch/wear-os',
            status: 'planned',
            priority: 'low'
          },
          {
            key: 'huawei-watch',
            icon: <TabletOutlined />,
            label: '华为手表',
            path: '/apps/watch/huawei',
            status: 'planned',
            priority: 'low'
          }
        ]
      },
      {
        key: 'web-apps',
        icon: <GlobalOutlined />,
        label: 'Web应用',
        path: '/apps/web',
        status: 'planned',
        description: '响应式Web、PWA、浏览器扩展',
        priority: 'high',
        children: [
          {
            key: 'responsive-web',
            icon: <GlobalOutlined />,
            label: '响应式Web',
            path: '/apps/web/responsive',
            status: 'completed',
            priority: 'high'
          },
          {
            key: 'pwa-app',
            icon: <RocketOutlined />,
            label: 'PWA应用',
            path: '/apps/web/pwa',
            status: 'planned',
            priority: 'medium'
          },
          {
            key: 'browser-extension',
            icon: <GlobalOutlined />,
            label: '浏览器扩展',
            path: '/apps/web/extension',
            status: 'planned',
            priority: 'low'
          }
        ]
      }
    ]
  }
];

/**
 * 个人中心菜单配置
 */
export const profileMenuConfig: MenuItem[] = [
  {
    key: 'profile-overview',
    icon: <UserOutlined />,
    label: '个人资料',
    path: '/profile',
    status: 'completed',
    priority: 'high'
  },
  {
    key: 'profile-settings',
    icon: <SettingOutlined />,
    label: '账户设置',
    path: '/profile/settings',
    status: 'completed',
    priority: 'high'
  },
  {
    key: 'profile-security',
    icon: <LockOutlined />,
    label: '安全设置',
    path: '/profile/security',
    status: 'completed',
    priority: 'high'
  },
  {
    key: 'profile-notifications',
    icon: <BellOutlined />,
    label: '通知设置',
    path: '/profile/notifications',
    status: 'completed',
    priority: 'medium'
  }
];

/**
 * 快捷操作菜单配置
 */
export const quickActionsConfig: MenuItem[] = [
  {
    key: 'quick-chat',
    icon: <MessageOutlined />,
    label: '快速对话',
    path: '/chat',
    status: 'completed',
    priority: 'high'
  },
  {
    key: 'quick-search',
    icon: <SearchOutlined />,
    label: '智慧搜索',
    path: '/wisdom/search',
    status: 'completed',
    priority: 'high'
  },
  {
    key: 'quick-note',
    icon: <EditOutlined />,
    label: '快速笔记',
    path: '/notes/new',
    status: 'completed',
    priority: 'medium'
  },
  {
    key: 'quick-task',
    icon: <CheckSquareOutlined />,
    label: '创建任务',
    path: '/projects/tasks/new',
    status: 'planned',
    priority: 'medium'
  }
];

/**
 * 根据用户角色和权限过滤菜单项
 */
export const filterMenuByPermissions = (
  menuItems: MenuItem[],
  userRoles: string[] = [],
  userPermissions: string[] = []
): MenuItem[] => {
  return menuItems.filter(item => {
    // 检查角色权限
    if (item.requiredRole && item.requiredRole.length > 0) {
      const hasRequiredRole = item.requiredRole.some(role => userRoles.includes(role));
      if (!hasRequiredRole) return false;
    }

    // 检查功能权限
    if (item.requiredPermission && item.requiredPermission.length > 0) {
      const hasRequiredPermission = item.requiredPermission.some(permission => 
        userPermissions.includes(permission)
      );
      if (!hasRequiredPermission) return false;
    }

    // 递归过滤子菜单
    if (item.children) {
      item.children = filterMenuByPermissions(item.children, userRoles, userPermissions);
    }

    return true;
  });
};

/**
 * 根据开发状态过滤菜单项
 */
export const filterMenuByStatus = (
  menuItems: MenuItem[],
  includeStatuses: ('completed' | 'partial' | 'planned')[] = ['completed', 'partial']
): MenuItem[] => {
  return menuItems.filter(item => {
    if (!includeStatuses.includes(item.status)) return false;

    // 递归过滤子菜单
    if (item.children) {
      item.children = filterMenuByStatus(item.children, includeStatuses);
    }

    return true;
  });
};

/**
 * 获取菜单项的状态徽章
 */
export const getStatusBadge = (status: MenuItem['status']): { text: string; color: string } => {
  switch (status) {
    case 'completed':
      return { text: '✅', color: '#52c41a' };
    case 'partial':
      return { text: '🔄', color: '#faad14' };
    case 'planned':
      return { text: '⏳', color: '#d9d9d9' };
    default:
      return { text: '', color: '' };
  }
};

/**
 * 获取优先级颜色
 */
export const getPriorityColor = (priority: MenuItem['priority']): string => {
  switch (priority) {
    case 'high':
      return '#ff4d4f';
    case 'medium':
      return '#faad14';
    case 'low':
      return '#52c41a';
    default:
      return '#d9d9d9';
  }
};

export default {
  mainMenuConfig,
  profileMenuConfig,
  quickActionsConfig,
  filterMenuByPermissions,
  filterMenuByStatus,
  getStatusBadge,
  getPriorityColor
};