// 权限管理模块全面测试脚本
// const fetch = require('node-fetch'); // 使用Node.js内置fetch

const BASE_URL = 'http://localhost:8080';
let adminToken = '';

// 测试配置
const testConfig = {
    adminCredentials: {
        username: 'admin',
        password: 'admin123'
    },
    testPermission: {
        name: `测试权限_${Date.now()}`,
        code: `test_permission_${Date.now()}`,
        description: '测试权限描述',
        resource: 'test_resource',
        action: 'test_action'
    },
    testRole: {
        name: `测试角色_${Date.now()}`,
        code: `test_role_${Date.now()}`,
        description: '测试角色描述',
        level: 1
    }
};

// 辅助函数
async function makeRequest(url, options = {}) {
    try {
        const response = await fetch(url, {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            },
            ...options
        });
        
        const data = await response.json();
        return { response, data };
    } catch (error) {
        console.error(`请求失败: ${error.message}`);
        return { error };
    }
}

// 测试函数
async function testAdminLogin() {
    console.log('\n--- 测试: 管理员登录 ---');
    console.log('1️⃣ 测试管理员登录...');
    
    const { response, data, error } = await makeRequest(`${BASE_URL}/api/v1/auth/login`, {
        method: 'POST',
        body: JSON.stringify(testConfig.adminCredentials)
    });
    
    if (error || !response.ok) {
        console.log('❌ 管理员登录失败');
        return false;
    }
    
    if (data.success && data.data.token) {
        adminToken = data.data.token;
        console.log('✅ 管理员登录成功');
        console.log(`   - Token长度: ${adminToken.length}`);
        return true;
    }
    
    console.log('❌ 管理员登录失败 - 无效响应');
    return false;
}

async function testCreatePermission() {
    console.log('\n--- 测试: 创建权限 ---');
    console.log('2️⃣ 测试创建权限...');
    
    const { response, data, error } = await makeRequest(`${BASE_URL}/api/v1/permissions`, {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${adminToken}`
        },
        body: JSON.stringify(testConfig.testPermission)
    });
    
    if (error) {
        console.log('❌ 创建权限失败 - 请求错误');
        return false;
    }
    
    if (response.ok && data.permission) {
        testConfig.testPermission.id = data.permission.id;
        console.log('✅ 创建权限成功');
        console.log(`   - 权限ID: ${data.permission.id}`);
        console.log(`   - 权限名称: ${data.permission.name}`);
        console.log(`   - 权限代码: ${data.permission.code}`);
        console.log(`   - 资源: ${data.permission.resource}`);
        console.log(`   - 操作: ${data.permission.action}`);
        console.log('✅ 创建权限 - 通过');
        return true;
    }
    
    console.log('❌ 创建权限失败');
    console.log(`   - 状态码: ${response.status}`);
    console.log(`   - 响应: ${JSON.stringify(data, null, 2)}`);
    console.log('❌ 创建权限 - 失败');
    return false;
}

async function testListPermissions() {
    console.log('\n--- 测试: 获取权限列表 ---');
    console.log('3️⃣ 测试获取权限列表...');
    
    const { response, data, error } = await makeRequest(`${BASE_URL}/api/v1/permissions`, {
        method: 'GET',
        headers: {
            'Authorization': `Bearer ${adminToken}`
        }
    });
    
    if (error) {
        console.log('❌ 获取权限列表失败 - 请求错误');
        return false;
    }
    
    if (response.ok && data.permissions) {
        console.log('✅ 获取权限列表成功');
        console.log(`   - 权限总数: ${data.total || data.permissions.length}`);
        console.log(`   - 当前页: ${data.page || 1}`);
        console.log(`   - 每页数量: ${data.limit || data.permissions.length}`);
        console.log('✅ 获取权限列表 - 通过');
        return true;
    }
    
    console.log('❌ 获取权限列表失败');
    console.log(`   - 状态码: ${response.status}`);
    console.log(`   - 响应: ${JSON.stringify(data, null, 2)}`);
    console.log('❌ 获取权限列表 - 失败');
    return false;
}

async function testGetPermissionById() {
    console.log('\n--- 测试: 根据ID获取权限 ---');
    console.log('4️⃣ 测试根据ID获取权限...');
    
    if (!testConfig.testPermission.id) {
        console.log('❌ 无测试权限ID');
        return false;
    }
    
    const { response, data, error } = await makeRequest(`${BASE_URL}/api/v1/permissions/${testConfig.testPermission.id}`, {
        method: 'GET',
        headers: {
            'Authorization': `Bearer ${adminToken}`
        }
    });
    
    if (error) {
        console.log('❌ 获取权限详情失败 - 请求错误');
        return false;
    }
    
    if (response.ok && data.success) {
        console.log('✅ 获取权限详情成功');
        console.log(`   - 权限名称: ${data.data.name}`);
        console.log(`   - 权限代码: ${data.data.code}`);
        console.log(`   - 资源: ${data.data.resource}`);
        console.log(`   - 操作: ${data.data.action}`);
        console.log('✅ 根据ID获取权限 - 通过');
        return true;
    }
    
    console.log('❌ 获取权限详情失败');
    console.log(`   - 状态码: ${response.status}`);
    console.log('❌ 根据ID获取权限 - 失败');
    return false;
}

async function testUpdatePermission() {
    console.log('\n--- 测试: 更新权限信息 ---');
    console.log('5️⃣ 测试更新权限信息...');
    
    if (!testConfig.testPermission.id) {
        console.log('❌ 无测试权限ID');
        return false;
    }
    
    const updateData = {
        description: '更新后的权限描述',
        resource: 'updated_resource'
    };
    
    const { response, data, error } = await makeRequest(`${BASE_URL}/api/v1/permissions/${testConfig.testPermission.id}`, {
        method: 'PUT',
        headers: {
            'Authorization': `Bearer ${adminToken}`
        },
        body: JSON.stringify(updateData)
    });
    
    if (error) {
        console.log('❌ 更新权限失败 - 请求错误');
        return false;
    }
    
    if (response.ok && data.success) {
        console.log('✅ 更新权限成功');
        console.log(`   - 新描述: ${data.data.description}`);
        console.log(`   - 新资源: ${data.data.resource}`);
        console.log('✅ 更新权限信息 - 通过');
        return true;
    }
    
    console.log('❌ 更新权限失败');
    console.log(`   - 状态码: ${response.status}`);
    console.log(`   - 响应: ${JSON.stringify(data, null, 2)}`);
    console.log('❌ 更新权限信息 - 失败');
    return false;
}

async function testCreateRole() {
    console.log('\n--- 测试: 创建角色 ---');
    console.log('6️⃣ 测试创建角色...');
    
    const { response, data, error } = await makeRequest(`${BASE_URL}/api/v1/roles`, {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${adminToken}`
        },
        body: JSON.stringify(testConfig.testRole)
    });
    
    if (error) {
        console.log('❌ 创建角色失败 - 请求错误');
        return false;
    }
    
    if (response.ok && data.role) {
        testConfig.testRole.id = data.role.id;
        console.log('✅ 创建角色成功');
        console.log(`   - 角色ID: ${data.role.id}`);
        console.log(`   - 角色名称: ${data.role.name}`);
        console.log(`   - 角色代码: ${data.role.code}`);
        console.log(`   - 角色等级: ${data.role.level}`);
        console.log(`   - 角色状态: ${data.role.status}`);
        console.log('✅ 创建角色 - 通过');
        return true;
    }
    
    console.log('❌ 创建角色失败');
    console.log(`   - 状态码: ${response.status}`);
    console.log(`   - 响应: ${JSON.stringify(data, null, 2)}`);
    console.log('❌ 创建角色 - 失败');
    return false;
}

async function testListRoles() {
    console.log('\n--- 测试: 获取角色列表 ---');
    console.log('7️⃣ 测试获取角色列表...');
    
    const { response, data, error } = await makeRequest(`${BASE_URL}/api/v1/roles`, {
        method: 'GET',
        headers: {
            'Authorization': `Bearer ${adminToken}`
        }
    });
    
    if (error) {
        console.log('❌ 获取角色列表失败 - 请求错误');
        return false;
    }
    
    if (response.ok && data.roles) {
        console.log('✅ 获取角色列表成功');
        console.log(`   - 角色总数: ${data.total || data.roles.length}`);
        console.log(`   - 当前页: ${data.page || 1}`);
        console.log(`   - 每页数量: ${data.limit || data.roles.length}`);
        console.log('✅ 获取角色列表 - 通过');
        return true;
    }
    
    console.log('❌ 获取角色列表失败');
    console.log(`   - 状态码: ${response.status}`);
    console.log(`   - 响应: ${JSON.stringify(data, null, 2)}`);
    console.log('❌ 获取角色列表 - 失败');
    return false;
}

async function testAssignPermissionToRole() {
    console.log('\n--- 测试: 为角色分配权限 ---');
    console.log('8️⃣ 测试为角色分配权限...');
    
    if (!testConfig.testRole.id || !testConfig.testPermission.id) {
        console.log('❌ 缺少测试角色ID或权限ID');
        return false;
    }
    
    const { response, data, error } = await makeRequest(`${BASE_URL}/api/v1/roles/${testConfig.testRole.id}/permissions`, {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${adminToken}`
        },
        body: JSON.stringify({
            permission_ids: [testConfig.testPermission.id]
        })
    });
    
    if (error) {
        console.log('❌ 分配权限失败 - 请求错误');
        return false;
    }
    
    if (response.ok && data.message && data.role_id) {
        console.log('✅ 分配权限成功');
        console.log(`   - 角色ID: ${data.role_id}`);
        console.log(`   - 分配的权限数量: ${data.permissions?.length || 0}`);
        console.log(`   - 消息: ${data.message}`);
        console.log('✅ 为角色分配权限 - 通过');
        return true;
    }
    
    console.log('❌ 分配权限失败');
    console.log(`   - 状态码: ${response.status}`);
    console.log(`   - 响应: ${JSON.stringify(data, null, 2)}`);
    console.log('❌ 为角色分配权限 - 失败');
    return false;
}

async function testGetRolePermissions() {
    console.log('\n--- 测试: 获取角色权限 ---');
    console.log('9️⃣ 测试获取角色权限...');
    
    if (!testConfig.testRole.id) {
        console.log('❌ 获取角色权限失败 - 缺少角色ID');
        return false;
    }
    
    const { response, data, error } = await makeRequest(`${BASE_URL}/api/v1/roles/${testConfig.testRole.id}/permissions`, {
        method: 'GET',
        headers: {
            'Authorization': `Bearer ${adminToken}`
        }
    });
    
    if (error) {
        console.log('❌ 获取角色权限失败 - 请求错误');
        return false;
    }
    
    if (response.ok && data.role_id && data.permissions) {
        console.log('✅ 获取角色权限成功');
        console.log(`   - 角色ID: ${data.role_id}`);
        console.log(`   - 角色名称: ${data.role_name}`);
        console.log(`   - 权限数量: ${data.permissions.length}`);
        console.log('✅ 获取角色权限 - 通过');
        return true;
    }
    
    console.log('❌ 获取角色权限失败');
    console.log(`   - 状态码: ${response.status}`);
    console.log(`   - 响应: ${JSON.stringify(data, null, 2)}`);
    console.log('❌ 获取角色权限 - 失败');
    return false;
}

async function testPermissionCheck() {
    console.log('\n--- 测试: 权限检查 ---');
    console.log('🔟 测试权限检查...');
    
    const checkRequest = {
        user_id: '1caed520-8bb5-4d86-b508-64fa0aebacfd', // admin用户ID
        resource: 'test_resource',
        action: 'test_action'
    };
    
    const { response, data, error } = await makeRequest(`${BASE_URL}/api/v1/permissions/check`, {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${adminToken}`
        },
        body: JSON.stringify(checkRequest)
    });
    
    if (error) {
        console.log('❌ 权限检查失败 - 请求错误');
        return false;
    }
    
    if (response.ok && data.hasOwnProperty('has_permission')) {
        console.log('✅ 权限检查成功');
        console.log(`   - 用户ID: ${data.user_id}`);
        console.log(`   - 资源: ${data.resource}`);
        console.log(`   - 操作: ${data.action}`);
        console.log(`   - 检查结果: ${data.has_permission ? '允许' : '拒绝'}`);
        console.log('✅ 权限检查 - 通过');
        return true;
    }
    
    console.log('❌ 权限检查失败');
    console.log(`   - 状态码: ${response.status}`);
    console.log(`   - 响应: ${JSON.stringify(data, null, 2)}`);
    console.log('❌ 权限检查 - 失败');
    return false;
}

async function testDeletePermission() {
    console.log('\n--- 测试: 删除权限 ---');
    console.log('1️⃣1️⃣ 测试删除权限...');
    
    if (!testConfig.testPermission.id) {
        console.log('❌ 无测试权限ID');
        return false;
    }
    
    const { response, data, error } = await makeRequest(`${BASE_URL}/api/v1/permissions/${testConfig.testPermission.id}`, {
        method: 'DELETE',
        headers: {
            'Authorization': `Bearer ${adminToken}`
        }
    });
    
    if (error) {
        console.log('❌ 删除权限失败 - 请求错误');
        return false;
    }
    
    if (response.ok && data.success) {
        console.log('✅ 删除权限成功');
        console.log('✅ 删除权限 - 通过');
        return true;
    }
    
    console.log('❌ 删除权限失败');
    console.log(`   - 状态码: ${response.status}`);
    console.log(`   - 响应: ${JSON.stringify(data, null, 2)}`);
    console.log('❌ 删除权限 - 失败');
    return false;
}

async function testDeleteRole() {
    console.log('\n--- 测试: 删除角色 ---');
    console.log('1️⃣2️⃣ 测试删除角色...');
    
    if (!testConfig.testRole.id) {
        console.log('❌ 无测试角色ID');
        return false;
    }
    
    const { response, data, error } = await makeRequest(`${BASE_URL}/api/v1/roles/${testConfig.testRole.id}`, {
        method: 'DELETE',
        headers: {
            'Authorization': `Bearer ${adminToken}`
        }
    });
    
    if (error) {
        console.log('❌ 删除角色失败 - 请求错误');
        return false;
    }
    
    if (response.ok && data.success) {
        console.log('✅ 删除角色成功');
        console.log('✅ 删除角色 - 通过');
        return true;
    }
    
    console.log('❌ 删除角色失败');
    console.log(`   - 状态码: ${response.status}`);
    console.log(`   - 响应: ${JSON.stringify(data, null, 2)}`);
    console.log('❌ 删除角色 - 失败');
    return false;
}

// 主测试函数
async function runPermissionManagementTests() {
    console.log('🚀 开始权限管理模块全面测试...\n');
    
    const tests = [
        { name: '管理员登录', fn: testAdminLogin },
        { name: '创建权限', fn: testCreatePermission },
        { name: '获取权限列表', fn: testListPermissions },
        { name: '根据ID获取权限', fn: testGetPermissionById },
        { name: '更新权限信息', fn: testUpdatePermission },
        { name: '创建角色', fn: testCreateRole },
        { name: '获取角色列表', fn: testListRoles },
        { name: '为角色分配权限', fn: testAssignPermissionToRole },
        { name: '获取角色权限', fn: testGetRolePermissions },
        { name: '权限检查', fn: testPermissionCheck },
        { name: '删除权限', fn: testDeletePermission },
        { name: '删除角色', fn: testDeleteRole }
    ];
    
    let passedTests = 0;
    let totalTests = tests.length;
    
    for (const test of tests) {
        try {
            const result = await test.fn();
            if (result) {
                passedTests++;
            }
        } catch (error) {
            console.log(`❌ ${test.name} - 异常: ${error.message}`);
        }
    }
    
    console.log('\n📊 权限管理模块测试结果汇总:');
    console.log(`✅ 通过: ${passedTests}/${totalTests}`);
    console.log(`❌ 失败: ${totalTests - passedTests}/${totalTests}`);
    
    if (passedTests === totalTests) {
        console.log('🎉 所有权限管理测试通过！');
    } else {
        console.log('⚠️  部分权限管理测试失败，请检查相关功能。');
    }
}

// 运行测试
runPermissionManagementTests().catch(console.error);