/**
 * 完整菜单数据创建脚本
 * 基于已开发的前端页面和后端API创建完整的菜单结构
 */

const completeMenuData = [
  // 1. 仪表板
  {
    id: 1,
    name: "仪表板",
    path: "/dashboard",
    icon: "DashboardOutlined",
    parent_id: null,
    sort_order: 1,
    status: "active",
    description: "系统概览、快捷操作、个性化推荐",
    permissions: ["dashboard:read"],
    roles: ["user", "admin"]
  },

  // 2. AI智能服务
  {
    id: 2,
    name: "AI智能服务",
    path: null,
    icon: "RobotOutlined",
    parent_id: null,
    sort_order: 2,
    status: "active",
    description: "AI对话、多模态AI、智能分析、内容生成",
    permissions: ["ai:read"],
    roles: ["user", "admin"]
  },
  {
    id: 21,
    name: "智能对话",
    path: "/chat",
    icon: "MessageOutlined",
    parent_id: 2,
    sort_order: 1,
    status: "active",
    description: "多轮对话、专业领域、语音交互",
    permissions: ["ai:chat"],
    roles: ["user", "admin"]
  },
  {
    id: 22,
    name: "多模态AI",
    path: "/ai/multimodal",
    icon: "CameraOutlined",
    parent_id: 2,
    sort_order: 2,
    status: "active",
    description: "图像生成、图像分析、视频处理、音频处理",
    permissions: ["ai:multimodal"],
    roles: ["user", "admin"]
  },
  {
    id: 221,
    name: "图像生成",
    path: "/ai/image-generation",
    icon: "PictureOutlined",
    parent_id: 22,
    sort_order: 1,
    status: "active",
    description: "AI图像创作和生成",
    permissions: ["ai:image:generate"],
    roles: ["user", "admin"]
  },
  {
    id: 222,
    name: "图像分析",
    path: "/ai/image-analysis",
    icon: "EyeOutlined",
    parent_id: 22,
    sort_order: 2,
    status: "active",
    description: "图像内容识别和分析",
    permissions: ["ai:image:analyze"],
    roles: ["user", "admin"]
  },
  {
    id: 23,
    name: "智能分析",
    path: "/ai/analysis",
    icon: "BarChartOutlined",
    parent_id: 2,
    sort_order: 3,
    status: "active",
    description: "数据分析、趋势预测、报告生成",
    permissions: ["ai:analysis"],
    roles: ["user", "admin"]
  },
  {
    id: 24,
    name: "内容生成",
    path: "/ai/generation",
    icon: "EditOutlined",
    parent_id: 2,
    sort_order: 4,
    status: "active",
    description: "文本生成、代码生成、创意设计",
    permissions: ["ai:generation"],
    roles: ["user", "admin"]
  },
  {
    id: 25,
    name: "AGI规划",
    path: "/ai/agi-planning",
    icon: "BulbOutlined",
    parent_id: 2,
    sort_order: 5,
    status: "active",
    description: "AGI智能规划和决策",
    permissions: ["ai:agi"],
    roles: ["user", "admin"]
  },
  {
    id: 26,
    name: "AGI推理",
    path: "/ai/agi-reasoning",
    icon: "ThunderboltOutlined",
    parent_id: 2,
    sort_order: 6,
    status: "active",
    description: "AGI智能推理和分析",
    permissions: ["ai:agi"],
    roles: ["user", "admin"]
  },
  {
    id: 27,
    name: "元学习",
    path: "/ai/meta-learning",
    icon: "ExperimentOutlined",
    parent_id: 2,
    sort_order: 7,
    status: "active",
    description: "AI元学习和自适应",
    permissions: ["ai:meta"],
    roles: ["user", "admin"]
  },
  {
    id: 28,
    name: "自我进化",
    path: "/ai/self-evolution",
    icon: "RiseOutlined",
    parent_id: 2,
    sort_order: 8,
    status: "active",
    description: "AI自我进化和优化",
    permissions: ["ai:evolution"],
    roles: ["user", "admin"]
  },

  // 3. 文化智慧
  {
    id: 3,
    name: "文化智慧",
    path: null,
    icon: "BookOutlined",
    parent_id: null,
    sort_order: 3,
    status: "active",
    description: "智慧库、搜索、推荐、分类管理",
    permissions: ["wisdom:read"],
    roles: ["user", "admin"]
  },
  {
    id: 31,
    name: "智慧库",
    path: "/wisdom",
    icon: "ReadOutlined",
    parent_id: 3,
    sort_order: 1,
    status: "active",
    description: "经典文献、现代解读、智慧问答",
    permissions: ["wisdom:read"],
    roles: ["user", "admin"]
  },
  {
    id: 32,
    name: "智慧详情",
    path: "/wisdom/detail",
    icon: "FileTextOutlined",
    parent_id: 3,
    sort_order: 2,
    status: "active",
    description: "智慧内容详细查看",
    permissions: ["wisdom:read"],
    roles: ["user", "admin"]
  },
  {
    id: 33,
    name: "推荐中心",
    path: "/recommendation",
    icon: "StarOutlined",
    parent_id: 3,
    sort_order: 3,
    status: "active",
    description: "个性化内容推荐",
    permissions: ["wisdom:recommend"],
    roles: ["user", "admin"]
  },

  // 4. 社区交流
  {
    id: 4,
    name: "社区交流",
    path: null,
    icon: "TeamOutlined",
    parent_id: null,
    sort_order: 4,
    status: "active",
    description: "社区动态、实时聊天、兴趣小组、活动中心",
    permissions: ["community:read"],
    roles: ["user", "admin"]
  },
  {
    id: 41,
    name: "社区动态",
    path: "/community",
    icon: "GlobalOutlined",
    parent_id: 4,
    sort_order: 1,
    status: "active",
    description: "社区最新动态和讨论",
    permissions: ["community:read"],
    roles: ["user", "admin"]
  },

  // 5. 智能学习
  {
    id: 5,
    name: "智能学习",
    path: null,
    icon: "BookOutlined",
    parent_id: null,
    sort_order: 5,
    status: "active",
    description: "课程中心、学习进度、能力评估、认证中心",
    permissions: ["learning:read"],
    roles: ["user", "admin"]
  },
  {
    id: 51,
    name: "课程中心",
    path: "/learning/courses",
    icon: "PlayCircleOutlined",
    parent_id: 5,
    sort_order: 1,
    status: "active",
    description: "学习课程浏览和播放",
    permissions: ["learning:course"],
    roles: ["user", "admin"]
  },
  {
    id: 52,
    name: "学习进度",
    path: "/learning/progress",
    icon: "ProgressOutlined",
    parent_id: 5,
    sort_order: 2,
    status: "active",
    description: "学习进度追踪和统计",
    permissions: ["learning:progress"],
    roles: ["user", "admin"]
  },
  {
    id: 53,
    name: "能力评估",
    path: "/learning/assessment",
    icon: "CheckCircleOutlined",
    parent_id: 5,
    sort_order: 3,
    status: "active",
    description: "技能测试和能力分析",
    permissions: ["learning:assessment"],
    roles: ["user", "admin"]
  },
  {
    id: 54,
    name: "智能学习",
    path: "/intelligent-learning",
    icon: "BulbOutlined",
    parent_id: 5,
    sort_order: 4,
    status: "active",
    description: "智能学习系统",
    permissions: ["learning:intelligent"],
    roles: ["user", "admin"]
  },
  {
    id: 55,
    name: "学习分析",
    path: "/learning-analytics",
    icon: "BarChartOutlined",
    parent_id: 5,
    sort_order: 5,
    status: "active",
    description: "学习数据分析仪表板",
    permissions: ["learning:analytics"],
    roles: ["user", "admin"]
  },
  {
    id: 56,
    name: "学习计划",
    path: "/learning-plan",
    icon: "CalendarOutlined",
    parent_id: 5,
    sort_order: 6,
    status: "active",
    description: "个性化学习计划",
    permissions: ["learning:plan"],
    roles: ["user", "admin"]
  },
  {
    id: 57,
    name: "成就中心",
    path: "/achievement",
    icon: "TrophyOutlined",
    parent_id: 5,
    sort_order: 7,
    status: "active",
    description: "学习成就和奖励",
    permissions: ["learning:achievement"],
    roles: ["user", "admin"]
  },

  // 6. 项目管理
  {
    id: 6,
    name: "项目管理",
    path: null,
    icon: "ProjectOutlined",
    parent_id: null,
    sort_order: 6,
    status: "active",
    description: "项目工作台、任务管理、团队协作、项目分析",
    permissions: ["project:read"],
    roles: ["user", "admin"]
  },
  {
    id: 61,
    name: "项目工作台",
    path: "/projects/workspace",
    icon: "DesktopOutlined",
    parent_id: 6,
    sort_order: 1,
    status: "active",
    description: "项目概览和工作台",
    permissions: ["project:workspace"],
    roles: ["user", "admin"]
  },
  {
    id: 62,
    name: "任务管理",
    path: "/projects/tasks",
    icon: "CheckSquareOutlined",
    parent_id: 6,
    sort_order: 2,
    status: "active",
    description: "任务创建、分配和跟踪",
    permissions: ["project:task"],
    roles: ["user", "admin"]
  },
  {
    id: 63,
    name: "团队协作",
    path: "/projects/collaboration",
    icon: "UsergroupAddOutlined",
    parent_id: 6,
    sort_order: 3,
    status: "active",
    description: "团队协作和沟通",
    permissions: ["project:collaboration"],
    roles: ["user", "admin"]
  },
  {
    id: 64,
    name: "项目分析",
    path: "/projects/analytics",
    icon: "LineChartOutlined",
    parent_id: 6,
    sort_order: 4,
    status: "active",
    description: "项目数据分析和报告",
    permissions: ["project:analytics"],
    roles: ["user", "admin"]
  },
  {
    id: 65,
    name: "项目管理",
    path: "/project-management",
    icon: "FolderOutlined",
    parent_id: 6,
    sort_order: 5,
    status: "active",
    description: "项目管理主页",
    permissions: ["project:manage"],
    roles: ["user", "admin"]
  },

  // 7. 健康管理
  {
    id: 7,
    name: "健康管理",
    path: null,
    icon: "HeartOutlined",
    parent_id: null,
    sort_order: 7,
    status: "active",
    description: "健康监测、健康档案、健康分析、健康建议",
    permissions: ["health:read"],
    roles: ["user", "admin"]
  },
  {
    id: 71,
    name: "健康监测",
    path: "/health/monitoring",
    icon: "MonitorOutlined",
    parent_id: 7,
    sort_order: 1,
    status: "active",
    description: "实时健康数据监测",
    permissions: ["health:monitor"],
    roles: ["user", "admin"]
  },
  {
    id: 72,
    name: "健康档案",
    path: "/health/records",
    icon: "FileProtectOutlined",
    parent_id: 7,
    sort_order: 2,
    status: "active",
    description: "个人健康档案管理",
    permissions: ["health:records"],
    roles: ["user", "admin"]
  },
  {
    id: 73,
    name: "健康分析",
    path: "/health/analysis",
    icon: "AreaChartOutlined",
    parent_id: 7,
    sort_order: 3,
    status: "active",
    description: "健康数据分析和趋势",
    permissions: ["health:analysis"],
    roles: ["user", "admin"]
  },
  {
    id: 74,
    name: "健康建议",
    path: "/health/advice",
    icon: "BulbOutlined",
    parent_id: 7,
    sort_order: 4,
    status: "active",
    description: "个性化健康建议",
    permissions: ["health:advice"],
    roles: ["user", "admin"]
  },
  {
    id: 75,
    name: "健康管理",
    path: "/health-management",
    icon: "MedicineBoxOutlined",
    parent_id: 7,
    sort_order: 5,
    status: "active",
    description: "健康管理主页",
    permissions: ["health:manage"],
    roles: ["user", "admin"]
  },

  // 8. 安全中心
  {
    id: 8,
    name: "安全中心",
    path: null,
    icon: "SafetyOutlined",
    parent_id: null,
    sort_order: 8,
    status: "active",
    description: "安全扫描、安全监控、安全教育、安全工具",
    permissions: ["security:read"],
    roles: ["user", "admin"]
  },
  {
    id: 81,
    name: "安全中心",
    path: "/security",
    icon: "ShieldOutlined",
    parent_id: 8,
    sort_order: 1,
    status: "active",
    description: "安全中心主页",
    permissions: ["security:read"],
    roles: ["user", "admin"]
  },

  // 9. 第三方集成
  {
    id: 9,
    name: "第三方集成",
    path: null,
    icon: "LinkOutlined",
    parent_id: null,
    sort_order: 9,
    status: "active",
    description: "插件管理、服务集成、Webhook管理、OAuth应用",
    permissions: ["integration:read"],
    roles: ["admin"]
  },
  {
    id: 91,
    name: "第三方集成",
    path: "/integration",
    icon: "ApiOutlined",
    parent_id: 9,
    sort_order: 1,
    status: "active",
    description: "第三方服务集成管理",
    permissions: ["integration:manage"],
    roles: ["admin"]
  },

  // 10. 系统管理
  {
    id: 10,
    name: "系统管理",
    path: null,
    icon: "SettingOutlined",
    parent_id: null,
    sort_order: 10,
    status: "active",
    description: "系统设置、用户管理、权限管理、系统监控",
    permissions: ["admin:read"],
    roles: ["admin"]
  },
  {
    id: 101,
    name: "管理仪表板",
    path: "/admin/dashboard",
    icon: "DashboardOutlined",
    parent_id: 10,
    sort_order: 1,
    status: "active",
    description: "管理员仪表板",
    permissions: ["admin:dashboard"],
    roles: ["admin"]
  },
  {
    id: 102,
    name: "用户管理",
    path: "/admin/users",
    icon: "UserOutlined",
    parent_id: 10,
    sort_order: 2,
    status: "active",
    description: "用户列表、用户权限、用户组",
    permissions: ["admin:user"],
    roles: ["admin"]
  },
  {
    id: 103,
    name: "角色管理",
    path: "/admin/roles",
    icon: "TeamOutlined",
    parent_id: 10,
    sort_order: 3,
    status: "active",
    description: "角色定义、角色权限、角色分配",
    permissions: ["admin:role"],
    roles: ["admin"]
  },
  {
    id: 104,
    name: "权限管理",
    path: "/admin/permissions",
    icon: "LockOutlined",
    parent_id: 10,
    sort_order: 4,
    status: "active",
    description: "权限定义、权限分配、访问控制",
    permissions: ["admin:permission"],
    roles: ["admin"]
  },
  {
    id: 105,
    name: "菜单管理",
    path: "/admin/menus",
    icon: "MenuOutlined",
    parent_id: 10,
    sort_order: 5,
    status: "active",
    description: "菜单配置、菜单权限、菜单排序",
    permissions: ["admin:menu"],
    roles: ["admin"]
  },
  {
    id: 106,
    name: "系统设置",
    path: "/admin/settings",
    icon: "ToolOutlined",
    parent_id: 10,
    sort_order: 6,
    status: "active",
    description: "基础配置、功能开关、性能调优",
    permissions: ["admin:setting"],
    roles: ["admin"]
  },
  {
    id: 107,
    name: "内容审核",
    path: "/admin/content-review",
    icon: "AuditOutlined",
    parent_id: 10,
    sort_order: 7,
    status: "active",
    description: "内容审核和管理",
    permissions: ["admin:content"],
    roles: ["admin"]
  },
  {
    id: 108,
    name: "分类管理",
    path: "/admin/categories",
    icon: "AppstoreOutlined",
    parent_id: 10,
    sort_order: 8,
    status: "active",
    description: "内容分类管理",
    permissions: ["admin:category"],
    roles: ["admin"]
  },
  {
    id: 109,
    name: "标签管理",
    path: "/admin/tags",
    icon: "TagsOutlined",
    parent_id: 10,
    sort_order: 9,
    status: "active",
    description: "标签管理和维护",
    permissions: ["admin:tag"],
    roles: ["admin"]
  },
  {
    id: 110,
    name: "智慧管理",
    path: "/admin/wisdom",
    icon: "BookOutlined",
    parent_id: 10,
    sort_order: 10,
    status: "active",
    description: "智慧内容管理",
    permissions: ["admin:wisdom"],
    roles: ["admin"]
  },
  {
    id: 111,
    name: "智慧编辑",
    path: "/admin/wisdom-editor",
    icon: "EditOutlined",
    parent_id: 10,
    sort_order: 11,
    status: "active",
    description: "智慧内容编辑",
    permissions: ["admin:wisdom:edit"],
    roles: ["admin"]
  },
  {
    id: 112,
    name: "数据分析",
    path: "/admin/analytics",
    icon: "BarChartOutlined",
    parent_id: 10,
    sort_order: 12,
    status: "active",
    description: "系统数据分析",
    permissions: ["admin:analytics"],
    roles: ["admin"]
  },
  {
    id: 113,
    name: "通知中心",
    path: "/admin/notifications",
    icon: "BellOutlined",
    parent_id: 10,
    sort_order: 13,
    status: "active",
    description: "系统通知管理",
    permissions: ["admin:notification"],
    roles: ["admin"]
  },
  {
    id: 114,
    name: "全球管理",
    path: "/admin/global",
    icon: "GlobalOutlined",
    parent_id: 10,
    sort_order: 14,
    status: "active",
    description: "全球化和本地化管理",
    permissions: ["admin:global"],
    roles: ["admin"]
  },

  // 11. 个人中心
  {
    id: 11,
    name: "个人中心",
    path: null,
    icon: "UserOutlined",
    parent_id: null,
    sort_order: 11,
    status: "active",
    description: "个人信息、账户设置、偏好配置",
    permissions: ["user:profile"],
    roles: ["user", "admin"]
  },
  {
    id: 111,
    name: "个人资料",
    path: "/profile",
    icon: "IdcardOutlined",
    parent_id: 11,
    sort_order: 1,
    status: "active",
    description: "个人信息管理",
    permissions: ["user:profile"],
    roles: ["user", "admin"]
  },
  {
    id: 112,
    name: "我的收藏",
    path: "/user/favorites",
    icon: "StarOutlined",
    parent_id: 11,
    sort_order: 2,
    status: "active",
    description: "收藏内容管理",
    permissions: ["user:favorites"],
    roles: ["user", "admin"]
  },
  {
    id: 113,
    name: "我的笔记",
    path: "/user/notes",
    icon: "FileTextOutlined",
    parent_id: 11,
    sort_order: 3,
    status: "active",
    description: "个人笔记管理",
    permissions: ["user:notes"],
    roles: ["user", "admin"]
  },
  {
    id: 114,
    name: "每日签到",
    path: "/daily-checkin",
    icon: "CalendarOutlined",
    parent_id: 11,
    sort_order: 4,
    status: "active",
    description: "每日签到和积分",
    permissions: ["user:checkin"],
    roles: ["user", "admin"]
  },

  // 12. 帮助中心
  {
    id: 12,
    name: "帮助中心",
    path: "/help",
    icon: "QuestionCircleOutlined",
    parent_id: null,
    sort_order: 12,
    status: "active",
    description: "使用帮助、常见问题、联系支持",
    permissions: ["help:read"],
    roles: ["user", "admin"]
  },

  // 13. 调试页面（开发环境）
  {
    id: 13,
    name: "调试页面",
    path: null,
    icon: "BugOutlined",
    parent_id: null,
    sort_order: 13,
    status: "active",
    description: "开发调试页面",
    permissions: ["debug:read"],
    roles: ["admin", "developer"]
  },
  {
    id: 131,
    name: "菜单调试",
    path: "/debug/menu",
    icon: "MenuOutlined",
    parent_id: 13,
    sort_order: 1,
    status: "active",
    description: "菜单系统调试",
    permissions: ["debug:menu"],
    roles: ["admin", "developer"]
  },
  {
    id: 132,
    name: "测试页面",
    path: "/test",
    icon: "ExperimentOutlined",
    parent_id: 13,
    sort_order: 2,
    status: "active",
    description: "功能测试页面",
    permissions: ["debug:test"],
    roles: ["admin", "developer"]
  },
  {
    id: 133,
    name: "简单测试",
    path: "/simple-test",
    icon: "PlayCircleOutlined",
    parent_id: 13,
    sort_order: 3,
    status: "active",
    description: "简单功能测试",
    permissions: ["debug:simple"],
    roles: ["admin", "developer"]
  },
  {
    id: 134,
    name: "Outlet测试",
    path: "/outlet-test",
    icon: "NodeIndexOutlined",
    parent_id: 13,
    sort_order: 4,
    status: "active",
    description: "路由Outlet测试",
    permissions: ["debug:outlet"],
    roles: ["admin", "developer"]
  }
];

// 创建菜单数据的SQL插入语句
function generateMenuSQL() {
  let sql = `-- 完整菜单数据插入脚本
-- 清空现有菜单数据
DELETE FROM menus;

-- 重置自增ID
ALTER SEQUENCE menus_id_seq RESTART WITH 1;

-- 插入菜单数据
INSERT INTO menus (id, name, path, icon, parent_id, sort_order, status, description, permissions, roles, created_at, updated_at) VALUES\n`;

  const values = completeMenuData.map(menu => {
    const permissions = JSON.stringify(menu.permissions);
    const roles = JSON.stringify(menu.roles);
    const path = menu.path ? `'${menu.path}'` : 'NULL';
    const parentId = menu.parent_id ? menu.parent_id : 'NULL';
    
    return `(${menu.id}, '${menu.name}', ${path}, '${menu.icon}', ${parentId}, ${menu.sort_order}, '${menu.status}', '${menu.description}', '${permissions}', '${roles}', NOW(), NOW())`;
  });

  sql += values.join(',\n');
  sql += ';\n\n-- 更新序列到正确的值\nSELECT setval(\'menus_id_seq\', (SELECT MAX(id) FROM menus));';
  
  return sql;
}

// 生成后端API调用脚本
async function testMenuAPI() {
  const baseURL = 'http://localhost:8080';
  
  try {
    console.log('🔄 正在测试菜单API...');
    
    // 1. 先尝试初始化菜单数据
    console.log('📝 初始化菜单数据...');
    const initResponse = await fetch(`${baseURL}/api/v1/menus/seed`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer test-token'
      }
    });
    
    if (initResponse.ok) {
      console.log('✅ 菜单数据初始化成功');
    } else {
      console.log('⚠️ 菜单数据初始化失败，可能已存在数据');
    }
    
    // 2. 获取菜单树
    console.log('🌳 获取菜单树...');
    const treeResponse = await fetch(`${baseURL}/api/v1/menus/tree`, {
      headers: {
        'Authorization': 'Bearer test-token'
      }
    });
    
    if (treeResponse.ok) {
      const menuTree = await treeResponse.json();
      console.log('✅ 菜单树获取成功');
      console.log('📊 菜单统计:');
      console.log(`   - 主菜单数量: ${menuTree.data?.length || 0}`);
      
      let totalMenus = 0;
      function countMenus(menus) {
        if (!menus) return 0;
        let count = menus.length;
        menus.forEach(menu => {
          if (menu.children) {
            count += countMenus(menu.children);
          }
        });
        return count;
      }
      
      totalMenus = countMenus(menuTree.data);
      console.log(`   - 总菜单数量: ${totalMenus}`);
      
      // 显示主菜单
      console.log('\n📋 主菜单列表:');
      menuTree.data?.forEach((menu, index) => {
        const childCount = menu.children ? menu.children.length : 0;
        console.log(`   ${index + 1}. ${menu.name} (${childCount} 个子菜单)`);
      });
      
    } else {
      console.log('❌ 菜单树获取失败');
    }
    
    // 3. 获取菜单列表
    console.log('\n📝 获取菜单列表...');
    const listResponse = await fetch(`${baseURL}/api/v1/menus?page=1&limit=50`, {
      headers: {
        'Authorization': 'Bearer test-token'
      }
    });
    
    if (listResponse.ok) {
      const menuList = await listResponse.json();
      console.log('✅ 菜单列表获取成功');
      console.log(`📊 列表统计: ${menuList.data?.total || 0} 个菜单项`);
    } else {
      console.log('❌ 菜单列表获取失败');
    }
    
    console.log('\n🎉 菜单API测试完成！');
    
  } catch (error) {
    console.error('❌ 菜单API测试失败:', error.message);
  }
}

// 输出SQL和测试
console.log('🚀 太上老君AI平台 - 完整菜单数据生成器');
console.log('=' .repeat(50));

console.log('\n📝 生成的SQL脚本:');
console.log(generateMenuSQL());

console.log('\n🔧 菜单数据统计:');
console.log(`- 总菜单项: ${completeMenuData.length}`);
console.log(`- 主菜单: ${completeMenuData.filter(m => !m.parent_id).length}`);
console.log(`- 子菜单: ${completeMenuData.filter(m => m.parent_id).length}`);

console.log('\n📋 主菜单概览:');
completeMenuData.filter(m => !m.parent_id).forEach((menu, index) => {
  const childCount = completeMenuData.filter(m => m.parent_id === menu.id).length;
  console.log(`${index + 1}. ${menu.name} (${childCount} 个子菜单)`);
});

// 运行API测试
testMenuAPI();