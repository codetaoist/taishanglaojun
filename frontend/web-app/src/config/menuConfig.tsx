/**
 * 太上老君AI平台 - 系统菜单配置
 * 基于系统菜单架构设计，集成所有功能模块
 */

import React from 'react';
import i18n from '../i18n';
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

  GlobalOutlined,
  ApiOutlined,
  BellOutlined,
  LockOutlined,
  MenuOutlined,
  DatabaseOutlined,
  ThunderboltOutlined,
  FireOutlined,
  RocketOutlined,
  AuditOutlined,
  AppstoreOutlined,
  KeyOutlined,
  LinkOutlined,
  SendOutlined
} from '@ant-design/icons';

export interface MenuItem {
  key: string;
  icon?: React.ReactNode;
  label: string;
  labelKey?: string;
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
    labelKey: 'mainMenu.labels.dashboard',
    path: '/dashboard',
    status: 'completed',
    description: '系统概览、快捷操作、个性化推荐',
    priority: 'high'
  },

  // 2. AI智能服务
  {
    key: 'ai-services',
    icon: <RobotOutlined />,
    label: 'AI智能服务',
    labelKey: 'mainMenu.labels.ai-services',
    status: 'partial',
    description: 'AI对话、多模态AI、智能分析、内容生成',
    requiredPermission: ['ai:read'],
    requiredRole: ['user', 'admin'],
    priority: 'high',
    children: [
      {
        key: 'ai-chat',
        icon: <MessageOutlined />,
        label: '智能对话',
        labelKey: 'mainMenu.labels.ai-chat',
        path: '/chat',
        status: 'completed',
        description: '多轮对话、专业领域、语音交互',
        priority: 'high'
      },
      {
        key: 'ai-multimodal',
        icon: <CameraOutlined />,
        label: '多模态AI',
        labelKey: 'mainMenu.labels.ai-multimodal',
        path: '/ai/multimodal',
        status: 'partial',
        description: '图像生成、图像分析、视频处理、音频处理',
        priority: 'high',
        children: [
          {
            key: 'image-generation',
            icon: <CameraOutlined />,
            label: '图像生成',
            labelKey: 'mainMenu.labels.image-generation',
            path: '/ai/multimodal/image-generation',
            status: 'planned',
            priority: 'medium'
          },
          {
            key: 'image-analysis',
            icon: <EyeOutlined />,
            label: '图像分析',
            labelKey: 'mainMenu.labels.image-analysis',
            path: '/ai/multimodal/image-analysis',
            status: 'planned',
            priority: 'medium'
          },
          {
            key: 'video-processing',
            icon: <VideoCameraOutlined />,
            label: '视频处理',
            labelKey: 'mainMenu.labels.video-processing',
            path: '/ai/multimodal/video-processing',
            status: 'planned',
            priority: 'low'
          },
          {
            key: 'audio-processing',
            icon: <SoundOutlined />,
            label: '音频处理',
            labelKey: 'mainMenu.labels.audio-processing',
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
        labelKey: 'mainMenu.labels.ai-analysis',
        path: '/ai/analysis',
        status: 'planned',
        description: '数据分析、趋势预测、报告生成',
        priority: 'medium',
        children: [
          {
            key: 'data-analysis',
            icon: <LineChartOutlined />,
            label: '数据分析',
            labelKey: 'mainMenu.labels.data-analysis',
            path: '/ai/analysis/data',
            status: 'planned',
            priority: 'medium'
          },
          {
            key: 'trend-prediction',
            icon: <ThunderboltOutlined />,
            label: '趋势预测',
            labelKey: 'mainMenu.labels.trend-prediction',
            path: '/ai/analysis/trends',
            status: 'planned',
            priority: 'medium'
          },
          {
            key: 'report-generation',
            icon: <FileTextOutlined />,
            label: '报告生成',
            labelKey: 'mainMenu.labels.report-generation',
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
        labelKey: 'mainMenu.labels.ai-generation',
        path: '/ai/generation',
        status: 'planned',
        description: '文本生成、代码生成、创意设计',
        priority: 'medium',
        children: [
          {
            key: 'text-generation',
            icon: <EditOutlined />,
            label: '文本生成',
            labelKey: 'mainMenu.labels.text-generation',
            path: '/ai/generation/text',
            status: 'planned',
            priority: 'medium'
          },
          {
            key: 'code-generation',
            icon: <ApiOutlined />,
            label: '代码生成',
            labelKey: 'mainMenu.labels.code-generation',
            path: '/ai/generation/code',
            status: 'planned',
            priority: 'medium'
          },
          {
            key: 'creative-design',
            icon: <FireOutlined />,
            label: '创意设计',
            labelKey: 'mainMenu.labels.creative-design',
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
    labelKey: 'mainMenu.labels.cultural-wisdom',
    status: 'completed',
    description: '智慧库、智慧搜索、智慧推荐、分类管理',
    priority: 'high',
    children: [
      {
        key: 'wisdom-library',
        icon: <BookOutlined />,
        label: '智慧库',
        labelKey: 'mainMenu.labels.wisdom-library',
        path: '/wisdom',
        status: 'completed',
        description: '经典文献、现代解读、智慧问答',
        priority: 'high'
      },
      {
        key: 'wisdom-search',
        icon: <SearchOutlined />,
        label: '智慧搜索',
        labelKey: 'mainMenu.labels.wisdom-search',
        path: '/wisdom/search',
        status: 'completed',
        description: '语义搜索、关联推荐、搜索历史',
        priority: 'high'
      },
      {
        key: 'wisdom-recommend',
        icon: <StarOutlined />,
        label: '智慧推荐',
        labelKey: 'mainMenu.labels.wisdom-recommend',
        path: '/recommendations',
        status: 'completed',
        description: '个性化推荐、每日智慧、主题推荐',
        priority: 'high'
      },
      {
        key: 'wisdom-category',
        icon: <TagOutlined />,
        label: '分类管理',
        labelKey: 'mainMenu.labels.wisdom-category',
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
        labelKey: 'mainMenu.labels.wisdom-favorites',
        path: '/favorites',
        status: 'completed',
        description: '收藏管理、个人智慧库',
        priority: 'medium'
      },
      {
        key: 'wisdom-notes',
        icon: <EditOutlined />,
        label: '我的笔记',
        labelKey: 'mainMenu.labels.wisdom-notes',
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
    labelKey: 'mainMenu.labels.community',
    status: 'completed',
    description: '社区动态、实时聊天、兴趣小组、活动中心',
    priority: 'high',
    children: [
      {
        key: 'community-posts',
        icon: <CommentOutlined />,
        label: '社区动态',
        labelKey: 'mainMenu.labels.community-posts',
        path: '/community',
        status: 'completed',
        description: '动态发布、互动功能、话题讨论',
        priority: 'high'
      },
      {
        key: 'community-chat',
        icon: <MessageOutlined />,
        label: '实时聊天',
        labelKey: 'mainMenu.labels.community-chat',
        path: '/community/chat',
        status: 'completed',
        description: '私聊功能、群聊功能、文件传输',
        priority: 'high'
      },
      {
        key: 'community-groups',
        icon: <UsergroupAddOutlined />,
        label: '兴趣小组',
        labelKey: 'mainMenu.labels.community-groups',
        path: '/community/groups',
        status: 'completed',
        description: '小组创建、成员管理、活动组织',
        priority: 'medium'
      },
      {
        key: 'community-events',
        icon: <CalendarOutlined />,
        label: '活动中心',
        labelKey: 'mainMenu.labels.community-events',
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
    labelKey: 'mainMenu.labels.intelligent-learning',
    status: 'partial',
    description: '课程中心、学习进度、能力评估、认证中心',
    priority: 'high',
    children: [
      {
        key: 'learning-courses',
        icon: <BookOutlined />,
        label: '课程中心',
        labelKey: 'mainMenu.labels.learning-courses',
        path: '/learning/courses',
        status: 'partial',
        description: '课程目录、课程播放、课程笔记',
        priority: 'high'
      },
      {
        key: 'learning-progress',
        icon: <LineChartOutlined />,
        label: '学习进度',
        labelKey: 'mainMenu.labels.learning-progress',
        path: '/learning/progress',
        status: 'completed',
        description: '进度追踪、学习计划、学习统计',
        priority: 'high'
      },
      {
        key: 'learning-assessment',
        icon: <ExperimentOutlined />,
        label: '能力评估',
        labelKey: 'mainMenu.labels.learning-assessment',
        path: '/learning/assessment',
        status: 'partial',
        description: '技能测试、能力分析、学习建议',
        priority: 'medium'
      },
      {
        key: 'learning-certificate',
        icon: <TrophyOutlined />,
        label: '认证中心',
        labelKey: 'mainMenu.labels.learning-certificate',
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
    labelKey: 'mainMenu.labels.project-management',
    status: 'completed',
    description: '项目工作台、任务管理、团队协作、项目分析',
    priority: 'medium',
    children: [
      {
        key: 'project-workspace',
        icon: <FolderOutlined />,
        label: '项目工作台',
        labelKey: 'mainMenu.labels.project-workspace',
        path: '/projects/workspace',
        status: 'completed',
        description: '项目概览、项目创建、项目模板',
        priority: 'medium'
      },
      {
        key: 'project-tasks',
        icon: <CheckSquareOutlined />,
        label: '任务管理',
        labelKey: 'mainMenu.labels.project-tasks',
        path: '/projects/tasks',
        status: 'completed',
        description: '任务创建、任务跟踪、任务提醒',
        priority: 'high'
      },
      {
        key: 'project-collaboration',
        icon: <UsergroupAddOutlined />,
        label: '团队协作',
        labelKey: 'mainMenu.labels.project-collaboration',
        path: '/projects/collaboration',
        status: 'completed',
        description: '团队管理、协作工具、文档共享',
        priority: 'medium'
      },
      {
        key: 'project-analytics',
        icon: <BarChartOutlined />,
        label: '项目分析',
        labelKey: 'mainMenu.labels.project-analytics',
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
    labelKey: 'mainMenu.labels.health-management',
    status: 'planned',
    description: '健康监测、健康分析、健康建议、健康档案',
    priority: 'high',
    children: [
      {
        key: 'health-monitor',
        icon: <MonitorOutlined />,
        label: '健康监测',
        labelKey: 'mainMenu.labels.health-monitor',
        path: '/health/monitor',
        status: 'planned',
        description: '生理监测、运动追踪、睡眠分析',
        priority: 'high'
      },
      {
        key: 'health-analysis',
        icon: <BarChartOutlined />,
        label: '健康分析',
        labelKey: 'mainMenu.labels.health-analysis',
        path: '/health/analysis',
        status: 'planned',
        description: '健康报告、趋势分析、异常预警',
        priority: 'high'
      },
      {
        key: 'health-advice',
        icon: <BulbOutlined />,
        label: '健康建议',
        labelKey: 'mainMenu.labels.health-advice',
        path: '/health/advice',
        status: 'planned',
        description: '个性化建议、运动计划、饮食建议',
        priority: 'medium'
      },
      {
        key: 'health-records',
        icon: <FileTextOutlined />,
        label: '健康档案',
        labelKey: 'mainMenu.labels.health-records',
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
    labelKey: 'mainMenu.labels.security-center',
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
        labelKey: 'mainMenu.labels.threat-detection',
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
        labelKey: 'mainMenu.labels.vulnerability-management',
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
        labelKey: 'mainMenu.labels.penetration-testing',
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
        labelKey: 'mainMenu.labels.security-education',
        path: '/security#security-education',
        status: 'completed',
        description: '安全培训课程、实验环境、认证管理',
        priority: 'medium'
      },
      {
        key: 'security-audit',
        icon: <AuditOutlined />,
        label: '安全审计',
        labelKey: 'mainMenu.labels.security-audit',
        path: '/security#security-audit',
        status: 'completed',
        description: '审计日志、安全事件、合规报告',
        priority: 'medium',
        requiredRole: ['admin', 'security']
      }
    ]
  },

  // 9. 第三方集成
  {
    key: 'third-party-integration',
    icon: <AppstoreOutlined />,
    label: '第三方集成',
    labelKey: 'mainMenu.labels.third-party-integration',
    status: 'completed',
    description: 'API管理、插件系统、服务集成、OAuth认证',
    priority: 'high',
    requiredRole: ['admin', 'developer'],
    children: [
      {
        key: 'api-keys',
        icon: <KeyOutlined />,
        label: 'API密钥管理',
        labelKey: 'mainMenu.labels.api-keys',
        path: '/integration',
        status: 'completed',
        description: 'API密钥创建、管理、使用统计',
        priority: 'high',
        requiredRole: ['admin', 'developer']
      },
      {
        key: 'plugins',
        icon: <AppstoreOutlined />,
        label: '插件管理',
        labelKey: 'mainMenu.labels.plugins',
        path: '/integration',
        status: 'completed',
        description: '插件安装、配置、启用、禁用',
        priority: 'high',
        requiredRole: ['admin']
      },
      {
        key: 'service-integration',
        icon: <LinkOutlined />,
        label: '服务集成',
        labelKey: 'mainMenu.labels.service-integration',
        path: '/integration',
        status: 'completed',
        description: '第三方服务集成配置和管理',
        priority: 'high',
        requiredRole: ['admin']
      },
      {
        key: 'webhooks',
        icon: <SendOutlined />,
        label: 'Webhook管理',
        labelKey: 'mainMenu.labels.webhooks',
        path: '/integration',
        status: 'completed',
        description: 'Webhook创建、测试、日志查看',
        priority: 'medium',
        requiredRole: ['admin', 'developer']
      },
      {
         key: 'oauth-apps',
         icon: <SafetyCertificateOutlined />,
         label: 'OAuth应用',
         labelKey: 'mainMenu.labels.oauth-apps',
         path: '/integration',
         status: 'completed',
         description: 'OAuth应用管理、令牌管理',
         priority: 'medium',
         requiredRole: ['admin']
       }
    ]
  },

  // 10. 接口文档
  {
    key: 'api-documentation',
    icon: <ApiOutlined />,
    label: '接口文档',
    labelKey: 'mainMenu.labels.api-documentation',
    status: 'completed',
    description: '接口目录、状态管理、版本控制、快速检索',
    priority: 'high',
    requiredRole: ['admin', 'developer'],
    children: [
      {
        key: 'api-catalog',
        icon: <BookOutlined />,
        label: '接口目录',
        labelKey: 'mainMenu.labels.api-catalog',
        path: '/api-docs/catalog',
        status: 'completed',
        description: '按功能模块分类展示所有接口',
        priority: 'high',
        requiredRole: ['admin', 'developer']
      },
      {
        key: 'api-status',
        icon: <MonitorOutlined />,
        label: '接口状态',
        labelKey: 'mainMenu.labels.api-status',
        path: '/api-docs/status',
        status: 'completed',
        description: '开发中、测试中、已上线状态管理',
        priority: 'high',
        requiredRole: ['admin', 'developer']
      },
      {
        key: 'api-versions',
        icon: <FileTextOutlined />,
        label: '版本管理',
        labelKey: 'mainMenu.labels.api-versions',
        path: '/api-docs/versions',
        status: 'completed',
        description: '记录接口变更历史和版本信息',
        priority: 'medium',
        requiredRole: ['admin', 'developer']
      },
      {
        key: 'api-search',
        icon: <SearchOutlined />,
        label: '快速检索',
        labelKey: 'mainMenu.labels.api-search',
        path: '/api-docs/search',
        status: 'completed',
        description: '支持关键词搜索和高级筛选',
        priority: 'medium',
        requiredRole: ['admin', 'developer']
      }
    ]
  },

  // 11. 系统管理
  {
    key: 'system-management',
    icon: <SettingOutlined />,
    label: '系统管理',
    labelKey: 'mainMenu.labels.system-management',
    status: 'completed',
    description: '系统设置、用户管理、权限管理、系统监控',
    priority: 'high',
    requiredRole: ['admin'],
    children: [
      {
        key: 'admin-dashboard',
        icon: <DashboardOutlined />,
        label: '管理员仪表板',
        labelKey: 'mainMenu.labels.admin-dashboard',
        path: '/admin/dashboard',
        status: 'completed',
        description: '用户统计、系统指标、内容概览',
        priority: 'high',
        requiredRole: ['admin']
      },
      {
        key: 'system-settings',
        icon: <SettingOutlined />,
        label: '系统设置',
        labelKey: 'mainMenu.labels.system-settings',
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
        labelKey: 'mainMenu.labels.system-users',
        path: '/admin/users',
        status: 'completed',
        description: '用户列表、用户权限、用户组',
        priority: 'high',
        requiredRole: ['admin']
      },
      {
        key: 'system-roles',
        icon: <TeamOutlined />,
        label: '角色管理',
        labelKey: 'mainMenu.labels.system-roles',
        path: '/admin/roles',
        status: 'completed',
        description: '角色定义、角色权限、角色分配',
        priority: 'high',
        requiredRole: ['admin'],
        requiredPermission: ['role:read']
      },
      {
        key: 'system-permissions',
        icon: <LockOutlined />,
        label: '权限管理',
        labelKey: 'mainMenu.labels.system-permissions',
        path: '/admin/permissions',
        status: 'completed',
        description: '权限定义、权限分配、访问控制',
        priority: 'high',
        requiredRole: ['admin'],
        requiredPermission: ['permission:read']
      },
      {
        key: 'system-menus',
        icon: <MenuOutlined />,
        label: '菜单管理',
        labelKey: 'mainMenu.labels.system-menus',
        path: '/admin/menus',
        status: 'completed',
        description: '菜单配置、菜单权限、菜单排序',
        priority: 'high',
        requiredRole: ['admin'],
        requiredPermission: ['menu:read']
      },
      {
        key: 'system-monitoring',
        icon: <MonitorOutlined />,
        label: '系统监控',
        labelKey: 'mainMenu.labels.system-monitoring',
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
        labelKey: 'mainMenu.labels.system-analytics',
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
        labelKey: 'mainMenu.labels.system-notifications',
        path: '/admin/notifications',
        status: 'completed',
        description: '系统通知、消息推送、通知模板',
        priority: 'medium',
        requiredRole: ['admin']
      }
      ,
      {
        key: 'system-database',
        icon: <DatabaseOutlined />,
        label: '数据库管理',
        labelKey: 'mainMenu.labels.system-database',
        path: '/admin/database',
        status: 'partial',
        description: '连接配置、表结构、只读查询、备份与恢复',
        priority: 'high',
        requiredRole: ['admin']
      },
      {
        key: 'system-logs',
        icon: <FileTextOutlined />,
        label: '日志管理',
        labelKey: 'mainMenu.labels.system-logs',
        path: '/admin/logs',
        status: 'completed',
        description: '用户操作、系统执行、错误日志、安全审计、性能日志',
        priority: 'high',
        requiredRole: ['admin']
      },
      {
        key: 'system-issues',
        icon: <AuditOutlined />,
        label: '问题跟踪',
        labelKey: 'mainMenu.labels.system-issues',
        path: '/admin/issues',
        status: 'completed',
        description: '实时监控、严重程度评估、解决方案建议、告警通知',
        priority: 'high',
        requiredRole: ['admin']
      }
    ]
  },


];

/**
 * 个人中心菜单配置
 */
export const profileMenuConfig: MenuItem[] = [
  {
    key: 'profile-overview',
    icon: <UserOutlined />,
    label: '个人资料',
    labelKey: 'mainMenu.labels.profile-overview',
    path: '/profile',
    status: 'completed',
    priority: 'high'
  },
  {
    key: 'profile-settings',
    icon: <SettingOutlined />,
    label: '账户设置',
    labelKey: 'mainMenu.labels.profile-settings',
    path: '/profile/settings',
    status: 'completed',
    priority: 'high'
  },
  {
    key: 'profile-security',
    icon: <LockOutlined />,
    label: '安全设置',
    labelKey: 'mainMenu.labels.profile-security',
    path: '/profile/security',
    status: 'completed',
    priority: 'high'
  },
  {
    key: 'profile-notifications',
    icon: <BellOutlined />,
    label: '通知设置',
    labelKey: 'mainMenu.labels.profile-notifications',
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
    labelKey: 'mainMenu.labels.quick-chat',
    path: '/chat',
    status: 'completed',
    priority: 'high'
  },
  {
    key: 'quick-search',
    icon: <SearchOutlined />,
    label: '智慧搜索',
    labelKey: 'mainMenu.labels.quick-search',
    path: '/wisdom/search',
    status: 'completed',
    priority: 'high'
  },
  {
    key: 'quick-note',
    icon: <EditOutlined />,
    label: '快速笔记',
    labelKey: 'mainMenu.labels.quick-note',
    path: '/notes/new',
    status: 'completed',
    priority: 'medium'
  },
  {
    key: 'quick-task',
    icon: <CheckSquareOutlined />,
    label: '创建任务',
    labelKey: 'mainMenu.labels.quick-task',
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
      const hasRequiredRole = item.requiredRole.some(role => 
        userRoles.some(userRole => userRole.toLowerCase() === role.toLowerCase())
      );
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

// 将标签初始化为当前语言的翻译（带安全回退）
const translateOrFallback = (lk: string, base?: string) => {
  const translated = i18n.t(lk);
  if (translated && translated !== lk) return translated;
  const safeBase = base && !String(base).startsWith('mainMenu.labels.') ? base : undefined;
  const tail = lk.startsWith('mainMenu.labels.') ? (lk.split('.').pop() || lk) : lk;
  return safeBase ?? tail;
};

const applyLocalization = (items: MenuItem[]) => {
  items.forEach(item => {
    const lk = item.labelKey || `mainMenu.labels.${item.key}`;
    item.labelKey = lk;
    item.label = translateOrFallback(lk, item.label as any);
    if (item.children && item.children.length > 0) {
      applyLocalization(item.children);
    }
  });
};

applyLocalization(mainMenuConfig);
applyLocalization(profileMenuConfig);
applyLocalization(quickActionsConfig);

// 语言切换时动态更新标签
i18n.on('languageChanged', () => {
  applyLocalization(mainMenuConfig);
  applyLocalization(profileMenuConfig);
  applyLocalization(quickActionsConfig);
});