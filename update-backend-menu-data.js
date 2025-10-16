/**
 * 通过后端API更新完整菜单数据
 */

const completeMenuData = [
  // 1. 仪表板
  {
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

  // 3. 文化智慧
  {
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

  // 4. 社区交流
  {
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

  // 5. 智能学习
  {
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

  // 6. 项目管理
  {
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

  // 7. 健康管理
  {
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

  // 8. 安全中心
  {
    name: "安全中心",
    path: "/security",
    icon: "SafetyOutlined",
    parent_id: null,
    sort_order: 8,
    status: "active",
    description: "安全扫描、安全监控、安全教育、安全工具",
    permissions: ["security:read"],
    roles: ["user", "admin"]
  },

  // 9. 第三方集成
  {
    name: "第三方集成",
    path: "/integration",
    icon: "LinkOutlined",
    parent_id: null,
    sort_order: 9,
    status: "active",
    description: "插件管理、服务集成、Webhook管理、OAuth应用",
    permissions: ["integration:read"],
    roles: ["admin"]
  },

  // 10. 系统管理
  {
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

  // 11. 个人中心
  {
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

  // 12. 帮助中心
  {
    name: "帮助中心",
    path: "/help",
    icon: "QuestionCircleOutlined",
    parent_id: null,
    sort_order: 12,
    status: "active",
    description: "使用帮助、常见问题、联系支持",
    permissions: ["help:read"],
    roles: ["user", "admin"]
  }
];

// 子菜单数据
const subMenuData = {
  "AI智能服务": [
    {
      name: "智能对话",
      path: "/chat",
      icon: "MessageOutlined",
      sort_order: 1,
      status: "active",
      description: "多轮对话、专业领域、语音交互",
      permissions: ["ai:chat"],
      roles: ["user", "admin"]
    },
    {
      name: "多模态AI",
      path: "/ai/multimodal",
      icon: "CameraOutlined",
      sort_order: 2,
      status: "active",
      description: "图像生成、图像分析、视频处理、音频处理",
      permissions: ["ai:multimodal"],
      roles: ["user", "admin"]
    },
    {
      name: "图像生成",
      path: "/ai/image-generation",
      icon: "PictureOutlined",
      sort_order: 3,
      status: "active",
      description: "AI图像创作和生成",
      permissions: ["ai:image:generate"],
      roles: ["user", "admin"]
    },
    {
      name: "图像分析",
      path: "/ai/image-analysis",
      icon: "EyeOutlined",
      sort_order: 4,
      status: "active",
      description: "图像内容识别和分析",
      permissions: ["ai:image:analyze"],
      roles: ["user", "admin"]
    },
    {
      name: "智能分析",
      path: "/ai/analysis",
      icon: "BarChartOutlined",
      sort_order: 5,
      status: "active",
      description: "数据分析、趋势预测、报告生成",
      permissions: ["ai:analysis"],
      roles: ["user", "admin"]
    },
    {
      name: "内容生成",
      path: "/ai/generation",
      icon: "EditOutlined",
      sort_order: 6,
      status: "active",
      description: "文本生成、代码生成、创意设计",
      permissions: ["ai:generation"],
      roles: ["user", "admin"]
    }
  ],
  "文化智慧": [
    {
      name: "智慧库",
      path: "/wisdom",
      icon: "ReadOutlined",
      sort_order: 1,
      status: "active",
      description: "经典文献、现代解读、智慧问答",
      permissions: ["wisdom:read"],
      roles: ["user", "admin"]
    },
    {
      name: "推荐中心",
      path: "/recommendation",
      icon: "StarOutlined",
      sort_order: 2,
      status: "active",
      description: "个性化内容推荐",
      permissions: ["wisdom:recommend"],
      roles: ["user", "admin"]
    }
  ],
  "社区交流": [
    {
      name: "社区动态",
      path: "/community",
      icon: "GlobalOutlined",
      sort_order: 1,
      status: "active",
      description: "社区最新动态和讨论",
      permissions: ["community:read"],
      roles: ["user", "admin"]
    }
  ],
  "智能学习": [
    {
      name: "课程中心",
      path: "/learning/courses",
      icon: "PlayCircleOutlined",
      sort_order: 1,
      status: "active",
      description: "学习课程浏览和播放",
      permissions: ["learning:course"],
      roles: ["user", "admin"]
    },
    {
      name: "学习进度",
      path: "/learning/progress",
      icon: "ProgressOutlined",
      sort_order: 2,
      status: "active",
      description: "学习进度追踪和统计",
      permissions: ["learning:progress"],
      roles: ["user", "admin"]
    },
    {
      name: "能力评估",
      path: "/learning/assessment",
      icon: "CheckCircleOutlined",
      sort_order: 3,
      status: "active",
      description: "技能测试和能力分析",
      permissions: ["learning:assessment"],
      roles: ["user", "admin"]
    },
    {
      name: "智能学习",
      path: "/intelligent-learning",
      icon: "BulbOutlined",
      sort_order: 4,
      status: "active",
      description: "智能学习系统",
      permissions: ["learning:intelligent"],
      roles: ["user", "admin"]
    }
  ],
  "项目管理": [
    {
      name: "项目工作台",
      path: "/projects/workspace",
      icon: "DesktopOutlined",
      sort_order: 1,
      status: "active",
      description: "项目概览和工作台",
      permissions: ["project:workspace"],
      roles: ["user", "admin"]
    },
    {
      name: "任务管理",
      path: "/projects/tasks",
      icon: "CheckSquareOutlined",
      sort_order: 2,
      status: "active",
      description: "任务创建、分配和跟踪",
      permissions: ["project:task"],
      roles: ["user", "admin"]
    },
    {
      name: "团队协作",
      path: "/projects/collaboration",
      icon: "UsergroupAddOutlined",
      sort_order: 3,
      status: "active",
      description: "团队协作和沟通",
      permissions: ["project:collaboration"],
      roles: ["user", "admin"]
    }
  ],
  "健康管理": [
    {
      name: "健康监测",
      path: "/health/monitoring",
      icon: "MonitorOutlined",
      sort_order: 1,
      status: "active",
      description: "实时健康数据监测",
      permissions: ["health:monitor"],
      roles: ["user", "admin"]
    },
    {
      name: "健康档案",
      path: "/health/records",
      icon: "FileProtectOutlined",
      sort_order: 2,
      status: "active",
      description: "个人健康档案管理",
      permissions: ["health:records"],
      roles: ["user", "admin"]
    },
    {
      name: "健康分析",
      path: "/health/analysis",
      icon: "AreaChartOutlined",
      sort_order: 3,
      status: "active",
      description: "健康数据分析和趋势",
      permissions: ["health:analysis"],
      roles: ["user", "admin"]
    }
  ],
  "系统管理": [
    {
      name: "管理仪表板",
      path: "/admin/dashboard",
      icon: "DashboardOutlined",
      sort_order: 1,
      status: "active",
      description: "管理员仪表板",
      permissions: ["admin:dashboard"],
      roles: ["admin"]
    },
    {
      name: "用户管理",
      path: "/admin/users",
      icon: "UserOutlined",
      sort_order: 2,
      status: "active",
      description: "用户列表、用户权限、用户组",
      permissions: ["admin:user"],
      roles: ["admin"]
    },
    {
      name: "角色管理",
      path: "/admin/roles",
      icon: "TeamOutlined",
      sort_order: 3,
      status: "active",
      description: "角色定义、角色权限、角色分配",
      permissions: ["admin:role"],
      roles: ["admin"]
    },
    {
      name: "权限管理",
      path: "/admin/permissions",
      icon: "LockOutlined",
      sort_order: 4,
      status: "active",
      description: "权限定义、权限分配、访问控制",
      permissions: ["admin:permission"],
      roles: ["admin"]
    },
    {
      name: "菜单管理",
      path: "/admin/menus",
      icon: "MenuOutlined",
      sort_order: 5,
      status: "active",
      description: "菜单配置、菜单权限、菜单排序",
      permissions: ["admin:menu"],
      roles: ["admin"]
    },
    {
      name: "系统设置",
      path: "/admin/settings",
      icon: "ToolOutlined",
      sort_order: 6,
      status: "active",
      description: "基础配置、功能开关、性能调优",
      permissions: ["admin:setting"],
      roles: ["admin"]
    }
  ],
  "个人中心": [
    {
      name: "个人资料",
      path: "/profile",
      icon: "IdcardOutlined",
      sort_order: 1,
      status: "active",
      description: "个人信息管理",
      permissions: ["user:profile"],
      roles: ["user", "admin"]
    },
    {
      name: "我的收藏",
      path: "/user/favorites",
      icon: "StarOutlined",
      sort_order: 2,
      status: "active",
      description: "收藏内容管理",
      permissions: ["user:favorites"],
      roles: ["user", "admin"]
    },
    {
      name: "我的笔记",
      path: "/user/notes",
      icon: "FileTextOutlined",
      sort_order: 3,
      status: "active",
      description: "个人笔记管理",
      permissions: ["user:notes"],
      roles: ["user", "admin"]
    }
  ]
};

async function updateMenuData() {
  const baseURL = 'http://localhost:8080';
  
  try {
    console.log('🚀 开始更新菜单数据...');
    
    // 1. 先获取现有菜单列表
    console.log('📋 获取现有菜单...');
    const listResponse = await fetch(`${baseURL}/api/v1/menus?page=1&limit=100`, {
      headers: {
        'Authorization': 'Bearer test-token'
      }
    });
    
    if (listResponse.ok) {
      const existingMenus = await listResponse.json();
      console.log(`📊 现有菜单数量: ${existingMenus.data?.items?.length || 0}`);
      
      // 2. 删除现有菜单（如果有的话）
      if (existingMenus.data?.items?.length > 0) {
        console.log('🗑️ 清理现有菜单...');
        for (const menu of existingMenus.data.items) {
          try {
            await fetch(`${baseURL}/api/v1/admin/menus/${menu.id}`, {
              method: 'DELETE',
              headers: {
                'Authorization': 'Bearer test-token'
              }
            });
          } catch (error) {
            console.log(`⚠️ 删除菜单 ${menu.name} 失败:`, error.message);
          }
        }
      }
    }
    
    // 3. 创建主菜单
    console.log('📝 创建主菜单...');
    const createdMainMenus = {};
    
    for (const menu of completeMenuData) {
      try {
        const response = await fetch(`${baseURL}/api/v1/admin/menus`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer test-token'
          },
          body: JSON.stringify(menu)
        });
        
        if (response.ok) {
          const result = await response.json();
          createdMainMenus[menu.name] = result.data;
          console.log(`✅ 创建主菜单: ${menu.name}`);
        } else {
          const error = await response.text();
          console.log(`❌ 创建主菜单失败 ${menu.name}:`, error);
        }
      } catch (error) {
        console.log(`❌ 创建主菜单异常 ${menu.name}:`, error.message);
      }
    }
    
    // 4. 创建子菜单
    console.log('📝 创建子菜单...');
    
    for (const [parentName, children] of Object.entries(subMenuData)) {
      const parentMenu = createdMainMenus[parentName];
      if (!parentMenu) {
        console.log(`⚠️ 找不到父菜单: ${parentName}`);
        continue;
      }
      
      for (const childMenu of children) {
        try {
          const menuData = {
            ...childMenu,
            parent_id: parentMenu.id
          };
          
          const response = await fetch(`${baseURL}/api/v1/admin/menus`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              'Authorization': 'Bearer test-token'
            },
            body: JSON.stringify(menuData)
          });
          
          if (response.ok) {
            console.log(`✅ 创建子菜单: ${parentName} -> ${childMenu.name}`);
          } else {
            const error = await response.text();
            console.log(`❌ 创建子菜单失败 ${parentName} -> ${childMenu.name}:`, error);
          }
        } catch (error) {
          console.log(`❌ 创建子菜单异常 ${parentName} -> ${childMenu.name}:`, error.message);
        }
      }
    }
    
    // 5. 验证最终结果
    console.log('🔍 验证菜单创建结果...');
    const finalResponse = await fetch(`${baseURL}/api/v1/menus/tree`, {
      headers: {
        'Authorization': 'Bearer test-token'
      }
    });
    
    if (finalResponse.ok) {
      const menuTree = await finalResponse.json();
      console.log('✅ 菜单树获取成功');
      console.log('📊 最终菜单统计:');
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
      console.log('\n📋 最终菜单结构:');
      menuTree.data?.forEach((menu, index) => {
        const childCount = menu.children ? menu.children.length : 0;
        console.log(`   ${index + 1}. ${menu.name} (${childCount} 个子菜单)`);
        if (menu.children && menu.children.length > 0) {
          menu.children.forEach((child, childIndex) => {
            console.log(`      ${childIndex + 1}. ${child.name}`);
          });
        }
      });
      
    } else {
      console.log('❌ 最终验证失败');
    }
    
    console.log('\n🎉 菜单数据更新完成！');
    
  } catch (error) {
    console.error('❌ 菜单数据更新失败:', error.message);
  }
}

// 运行更新
updateMenuData();