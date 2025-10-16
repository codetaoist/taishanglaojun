/**
 * 角色管理测试数据添加脚本
 * 用于为角色管理功能添加各种类型的测试数据
 */

const API_BASE_URL = 'http://localhost:8080/api/v1';

// 测试角色数据
const testRoles = [
  {
    name: '内容管理员',
    code: 'content_manager',
    description: '负责网站内容的创建、编辑和发布，管理文章分类和标签，审核用户提交的内容',
    type: 'functional',
    level: 3,
    is_active: true,
    permissions: ['content_read', 'content_write', 'content_publish']
  },
  {
    name: '用户运营',
    code: 'user_operator',
    description: '负责用户管理、用户行为分析、用户反馈处理，维护良好的用户体验',
    type: 'functional',
    level: 2,
    is_active: true,
    permissions: ['user_read', 'user_write', 'analytics_read']
  },
  {
    name: '数据分析师',
    code: 'data_analyst',
    description: '负责数据收集、分析和报告生成，为业务决策提供数据支持',
    type: 'data',
    level: 4,
    is_active: true,
    permissions: ['analytics_read', 'analytics_write', 'report_generate']
  },
  {
    name: '客服专员',
    code: 'customer_service',
    description: '处理用户咨询、投诉和建议，提供技术支持和问题解决方案',
    type: 'functional',
    level: 1,
    is_active: true,
    permissions: ['user_read', 'ticket_read', 'ticket_write']
  },
  {
    name: '财务审计',
    code: 'financial_auditor',
    description: '负责财务数据审核、合规检查和风险控制，确保财务安全',
    type: 'data',
    level: 5,
    is_active: true,
    permissions: ['finance_read', 'audit_read', 'audit_write']
  },
  {
    name: '产品经理',
    code: 'product_manager',
    description: '负责产品规划、需求分析和功能设计，协调开发团队实现产品目标',
    type: 'custom',
    level: 4,
    is_active: true,
    permissions: ['product_read', 'product_write', 'analytics_read', 'user_read']
  },
  {
    name: '测试工程师',
    code: 'qa_engineer',
    description: '负责软件测试、质量保证和缺陷跟踪，确保产品质量',
    type: 'functional',
    level: 3,
    is_active: true,
    permissions: ['test_read', 'test_write', 'bug_report']
  },
  {
    name: '市场推广',
    code: 'marketing_specialist',
    description: '负责市场活动策划、品牌推广和用户增长，提升产品知名度',
    type: 'custom',
    level: 2,
    is_active: true,
    permissions: ['marketing_read', 'marketing_write', 'analytics_read']
  },
  {
    name: '安全专员',
    code: 'security_officer',
    description: '负责系统安全监控、漏洞修复和安全策略制定，保障系统安全',
    type: 'system',
    level: 5,
    is_active: true,
    permissions: ['security_read', 'security_write', 'audit_read', 'system_monitor']
  },
  {
    name: '临时访客',
    code: 'temp_visitor',
    description: '临时访问权限，用于短期合作伙伴或外部顾问的有限访问',
    type: 'custom',
    level: 1,
    is_active: false,
    permissions: ['basic_read']
  }
];

// 测试权限数据
const testPermissions = [
  {
    name: 'content_read',
    description: '查看内容',
    resource: 'content',
    action: 'read'
  },
  {
    name: 'content_write',
    description: '编辑内容',
    resource: 'content',
    action: 'write'
  },
  {
    name: 'content_publish',
    description: '发布内容',
    resource: 'content',
    action: 'publish'
  },
  {
    name: 'user_read',
    description: '查看用户信息',
    resource: 'user',
    action: 'read'
  },
  {
    name: 'user_write',
    description: '编辑用户信息',
    resource: 'user',
    action: 'write'
  },
  {
    name: 'analytics_read',
    description: '查看分析数据',
    resource: 'analytics',
    action: 'read'
  },
  {
    name: 'analytics_write',
    description: '编辑分析配置',
    resource: 'analytics',
    action: 'write'
  },
  {
    name: 'report_generate',
    description: '生成报告',
    resource: 'report',
    action: 'generate'
  },
  {
    name: 'ticket_read',
    description: '查看工单',
    resource: 'ticket',
    action: 'read'
  },
  {
    name: 'ticket_write',
    description: '处理工单',
    resource: 'ticket',
    action: 'write'
  },
  {
    name: 'finance_read',
    description: '查看财务数据',
    resource: 'finance',
    action: 'read'
  },
  {
    name: 'audit_read',
    description: '查看审计日志',
    resource: 'audit',
    action: 'read'
  },
  {
    name: 'audit_write',
    description: '编辑审计配置',
    resource: 'audit',
    action: 'write'
  },
  {
    name: 'product_read',
    description: '查看产品信息',
    resource: 'product',
    action: 'read'
  },
  {
    name: 'product_write',
    description: '编辑产品信息',
    resource: 'product',
    action: 'write'
  },
  {
    name: 'test_read',
    description: '查看测试结果',
    resource: 'test',
    action: 'read'
  },
  {
    name: 'test_write',
    description: '执行测试',
    resource: 'test',
    action: 'write'
  },
  {
    name: 'bug_report',
    description: '报告缺陷',
    resource: 'bug',
    action: 'report'
  },
  {
    name: 'marketing_read',
    description: '查看营销数据',
    resource: 'marketing',
    action: 'read'
  },
  {
    name: 'marketing_write',
    description: '编辑营销活动',
    resource: 'marketing',
    action: 'write'
  },
  {
    name: 'security_read',
    description: '查看安全日志',
    resource: 'security',
    action: 'read'
  },
  {
    name: 'security_write',
    description: '配置安全策略',
    resource: 'security',
    action: 'write'
  },
  {
    name: 'system_monitor',
    description: '系统监控',
    resource: 'system',
    action: 'monitor'
  },
  {
    name: 'basic_read',
    description: '基础查看权限',
    resource: 'basic',
    action: 'read'
  }
];

// HTTP 请求函数
async function makeRequest(url, method = 'GET', data = null) {
  const options = {
    method,
    headers: {
      'Content-Type': 'application/json',
    },
  };

  if (data) {
    options.body = JSON.stringify(data);
  }

  try {
    const response = await fetch(url, options);
    const result = await response.json();
    
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${result.message || response.statusText}`);
    }
    
    return result;
  } catch (error) {
    console.error(`请求失败 ${method} ${url}:`, error.message);
    throw error;
  }
}

// 添加权限数据
async function addPermissions() {
  console.log('🔐 开始添加权限数据...');
  
  for (const permission of testPermissions) {
    try {
      await makeRequest(`${API_BASE_URL}/permissions`, 'POST', permission);
      console.log(`✅ 权限添加成功: ${permission.name}`);
    } catch (error) {
      if (error.message.includes('already exists') || error.message.includes('重复')) {
        console.log(`⚠️  权限已存在: ${permission.name}`);
      } else {
        console.error(`❌ 权限添加失败: ${permission.name} - ${error.message}`);
      }
    }
  }
}

// 添加角色数据
async function addRoles() {
  console.log('\n👥 开始添加角色数据...');
  
  for (const role of testRoles) {
    try {
      await makeRequest(`${API_BASE_URL}/roles`, 'POST', role);
      console.log(`✅ 角色添加成功: ${role.name} (${role.code})`);
    } catch (error) {
      if (error.message.includes('already exists') || error.message.includes('重复')) {
        console.log(`⚠️  角色已存在: ${role.name} (${role.code})`);
      } else {
        console.error(`❌ 角色添加失败: ${role.name} - ${error.message}`);
      }
    }
  }
}

// 获取现有数据统计
async function getDataStats() {
  try {
    const rolesResponse = await makeRequest(`${API_BASE_URL}/roles`);
    const permissionsResponse = await makeRequest(`${API_BASE_URL}/permissions`);
    
    const roles = rolesResponse.data || [];
    const permissions = permissionsResponse.data || [];
    
    console.log('\n📊 数据统计:');
    console.log(`   角色总数: ${roles.length}`);
    console.log(`   权限总数: ${permissions.length}`);
    
    // 按类型统计角色
    const rolesByType = roles.reduce((acc, role) => {
      acc[role.type] = (acc[role.type] || 0) + 1;
      return acc;
    }, {});
    
    console.log('\n   角色类型分布:');
    Object.entries(rolesByType).forEach(([type, count]) => {
      console.log(`     ${type}: ${count}`);
    });
    
    // 按状态统计角色
    const activeRoles = roles.filter(role => role.is_active).length;
    const inactiveRoles = roles.length - activeRoles;
    
    console.log('\n   角色状态分布:');
    console.log(`     启用: ${activeRoles}`);
    console.log(`     禁用: ${inactiveRoles}`);
    
  } catch (error) {
    console.error('❌ 获取数据统计失败:', error.message);
  }
}

// 主函数
async function main() {
  console.log('🚀 开始添加角色管理测试数据...\n');
  
  try {
    // 跳过健康检查，直接尝试添加数据
    console.log('📝 开始添加测试数据（跳过健康检查）...\n');
    
    // 添加权限数据
    await addPermissions();
    
    // 添加角色数据
    await addRoles();
    
    // 显示统计信息
    await getDataStats();
    
    console.log('\n🎉 测试数据添加完成！');
    console.log('\n💡 提示:');
    console.log('   - 可以访问 http://localhost:5173/admin/roles 查看角色管理页面');
    console.log('   - 测试数据包含了不同类型、级别和状态的角色');
    console.log('   - 可以测试搜索、筛选、编辑、删除等功能');
    
  } catch (error) {
    console.error('\n❌ 测试数据添加失败:', error.message);
    console.log('\n🔧 故障排除:');
    console.log('   1. 确保后端服务正在运行 (http://localhost:8080)');
    console.log('   2. 检查数据库连接是否正常');
    console.log('   3. 确认API端点是否正确');
    console.log('   4. 检查是否需要授权头');
    process.exit(1);
  }
}

// 检查运行环境
if (typeof window !== 'undefined') {
    console.error('此脚本需要在Node.js环境中运行');
    process.exit(1);
}

// 使用Node.js内置的fetch API (Node.js 18+)
if (typeof fetch === 'undefined') {
    console.error('此脚本需要Node.js 18或更高版本（支持内置fetch API）');
    process.exit(1);
}

// 运行主函数
main();

module.exports = {
  testRoles,
  testPermissions,
  addPermissions,
  addRoles,
  getDataStats
};