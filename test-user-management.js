// 用户管理模块全面测试脚本
// const fetch = require('node-fetch'); // 使用Node.js内置fetch

const BASE_URL = 'http://localhost:8080';
let adminToken = '';

// 测试配置
const testConfig = {
    adminCredentials: {
        username: 'admin',
        password: 'admin123'
    },
    testUser: {
        username: `testuser_${Date.now()}`,
        email: `test_${Date.now()}@example.com`,
        password: 'testpass123',
        role: 'USER'
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

async function testCreateUser() {
    console.log('\n--- 测试: 创建用户 ---');
    console.log('2️⃣ 测试创建用户...');
    
    const { response, data, error } = await makeRequest(`${BASE_URL}/api/v1/admin/users`, {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${adminToken}`
        },
        body: JSON.stringify(testConfig.testUser)
    });
    
    if (error) {
        console.log('❌ 创建用户失败 - 请求错误');
        return false;
    }
    
    if (response.ok && data.success) {
        testConfig.testUser.id = data.data.id;
        console.log('✅ 创建用户成功');
        console.log(`   - 用户ID: ${data.data.id}`);
        console.log(`   - 用户名: ${data.data.username}`);
        console.log(`   - 邮箱: ${data.data.email}`);
        console.log(`   - 角色: ${data.data.role}`);
        console.log('✅ 创建用户 - 通过');
        return true;
    }
    
    console.log('❌ 创建用户失败');
    console.log(`   - 状态码: ${response.status}`);
    console.log(`   - 响应: ${JSON.stringify(data, null, 2)}`);
    console.log('❌ 创建用户 - 失败');
    return false;
}

async function testGetUserList() {
    console.log('\n--- 测试: 获取用户列表 ---');
    console.log('3️⃣ 测试获取用户列表...');
    
    const { response, data, error } = await makeRequest(`${BASE_URL}/api/v1/admin/users?page=1&limit=10`, {
        method: 'GET',
        headers: {
            'Authorization': `Bearer ${adminToken}`
        }
    });
    
    if (error) {
        console.log('❌ 获取用户列表失败 - 请求错误');
        return false;
    }
    
    if (response.ok && data.success) {
        console.log('✅ 获取用户列表成功');
        console.log(`   - 总用户数: ${data.data.total}`);
        console.log(`   - 当前页用户数: ${data.data.users.length}`);
        console.log(`   - 页码: ${data.data.page}`);
        console.log(`   - 每页数量: ${data.data.limit}`);
        console.log('✅ 获取用户列表 - 通过');
        return true;
    }
    
    console.log('❌ 获取用户列表失败');
    console.log(`   - 状态码: ${response.status}`);
    console.log('❌ 获取用户列表 - 失败');
    return false;
}

async function testGetUserById() {
    console.log('\n--- 测试: 根据ID获取用户 ---');
    console.log('4️⃣ 测试根据ID获取用户...');
    
    if (!testConfig.testUser.id) {
        console.log('❌ 无测试用户ID');
        return false;
    }
    
    const { response, data, error } = await makeRequest(`${BASE_URL}/api/v1/admin/users/${testConfig.testUser.id}`, {
        method: 'GET',
        headers: {
            'Authorization': `Bearer ${adminToken}`
        }
    });
    
    if (error) {
        console.log('❌ 获取用户详情失败 - 请求错误');
        return false;
    }
    
    if (response.ok && data.success) {
        console.log('✅ 获取用户详情成功');
        console.log(`   - 用户名: ${data.data.username}`);
        console.log(`   - 邮箱: ${data.data.email}`);
        console.log(`   - 角色: ${data.data.role}`);
        console.log(`   - 状态: ${data.data.status}`);
        console.log('✅ 根据ID获取用户 - 通过');
        return true;
    }
    
    console.log('❌ 获取用户详情失败');
    console.log(`   - 状态码: ${response.status}`);
    console.log('❌ 根据ID获取用户 - 失败');
    return false;
}

async function testUpdateUser() {
    console.log('\n--- 测试: 更新用户信息 ---');
    console.log('5️⃣ 测试更新用户信息...');
    
    if (!testConfig.testUser.id) {
        console.log('❌ 无测试用户ID');
        return false;
    }
    
    const updateData = {
        email: `updated_${Date.now()}@example.com`,
        role: 'MODERATOR'
    };
    
    const { response, data, error } = await makeRequest(`${BASE_URL}/api/v1/admin/users/${testConfig.testUser.id}`, {
        method: 'PUT',
        headers: {
            'Authorization': `Bearer ${adminToken}`
        },
        body: JSON.stringify(updateData)
    });
    
    if (error) {
        console.log('❌ 更新用户失败 - 请求错误');
        return false;
    }
    
    if (response.ok && data.success) {
        console.log('✅ 更新用户成功');
        console.log(`   - 新邮箱: ${data.data.email}`);
        console.log(`   - 新角色: ${data.data.role}`);
        console.log('✅ 更新用户信息 - 通过');
        return true;
    }
    
    console.log('❌ 更新用户失败');
    console.log(`   - 状态码: ${response.status}`);
    console.log(`   - 响应: ${JSON.stringify(data, null, 2)}`);
    console.log('❌ 更新用户信息 - 失败');
    return false;
}

async function testSearchUsers() {
    console.log('\n--- 测试: 搜索用户 ---');
    console.log('6️⃣ 测试搜索用户...');
    
    const { response, data, error } = await makeRequest(`${BASE_URL}/api/v1/admin/users?search=admin`, {
        method: 'GET',
        headers: {
            'Authorization': `Bearer ${adminToken}`
        }
    });
    
    if (error) {
        console.log('❌ 搜索用户失败 - 请求错误');
        return false;
    }
    
    if (response.ok && data.success) {
        console.log('✅ 搜索用户成功');
        console.log(`   - 搜索结果数: ${data.data.length}`);
        if (data.data.length > 0) {
            console.log(`   - 第一个结果: ${data.data[0].username}`);
        }
        console.log('✅ 搜索用户 - 通过');
        return true;
    }
    
    console.log('❌ 搜索用户失败');
    console.log(`   - 状态码: ${response.status}`);
    console.log('❌ 搜索用户 - 失败');
    return false;
}

async function testUserStatusToggle() {
    console.log('\n--- 测试: 用户状态切换 ---');
    console.log('7️⃣ 测试用户状态切换...');
    
    if (!testConfig.testUser.id) {
        console.log('❌ 无测试用户ID');
        return false;
    }
    
    // 禁用用户
    const { response: disableResponse, data: disableData, error: disableError } = await makeRequest(`${BASE_URL}/api/v1/admin/users/${testConfig.testUser.id}/status`, {
        method: 'PUT',
        headers: {
            'Authorization': `Bearer ${adminToken}`
        },
        body: JSON.stringify({ status: 'inactive' })
    });
    
    if (disableError || !disableResponse.ok) {
        console.log('❌ 禁用用户失败');
        return false;
    }
    
    // 启用用户
    const { response: enableResponse, data: enableData, error: enableError } = await makeRequest(`${BASE_URL}/api/v1/admin/users/${testConfig.testUser.id}/status`, {
        method: 'PUT',
        headers: {
            'Authorization': `Bearer ${adminToken}`
        },
        body: JSON.stringify({ status: 'active' })
    });
    
    if (enableError || !enableResponse.ok) {
        console.log('❌ 启用用户失败');
        return false;
    }
    
    console.log('✅ 用户状态切换成功');
    console.log('✅ 用户状态切换 - 通过');
    return true;
}

async function testDeleteUser() {
    console.log('\n--- 测试: 删除用户 ---');
    console.log('8️⃣ 测试删除用户...');
    
    if (!testConfig.testUser.id) {
        console.log('❌ 无测试用户ID');
        return false;
    }
    
    const { response, data, error } = await makeRequest(`${BASE_URL}/api/v1/admin/users/${testConfig.testUser.id}`, {
        method: 'DELETE',
        headers: {
            'Authorization': `Bearer ${adminToken}`
        }
    });
    
    if (error) {
        console.log('❌ 删除用户失败 - 请求错误');
        return false;
    }
    
    if (response.ok && data.success) {
        console.log('✅ 删除用户成功');
        console.log('✅ 删除用户 - 通过');
        return true;
    }
    
    console.log('❌ 删除用户失败');
    console.log(`   - 状态码: ${response.status}`);
    console.log(`   - 响应: ${JSON.stringify(data, null, 2)}`);
    console.log('❌ 删除用户 - 失败');
    return false;
}

async function testVerifyUserDeleted() {
    console.log('\n--- 测试: 验证用户已删除 ---');
    console.log('9️⃣ 验证用户已删除...');
    
    if (!testConfig.testUser.id) {
        console.log('❌ 无测试用户ID');
        return false;
    }
    
    const { response, data, error } = await makeRequest(`${BASE_URL}/api/v1/admin/users/${testConfig.testUser.id}`, {
        method: 'GET',
        headers: {
            'Authorization': `Bearer ${adminToken}`
        }
    });
    
    if (error) {
        console.log('❌ 验证删除失败 - 请求错误');
        return false;
    }
    
    if (response.status === 404) {
        console.log('✅ 用户已成功删除');
        console.log('✅ 验证用户已删除 - 通过');
        return true;
    }
    
    console.log('❌ 用户删除验证失败 - 用户仍然存在');
    console.log('❌ 验证用户已删除 - 失败');
    return false;
}

// 主测试函数
async function runUserManagementTests() {
    console.log('🚀 开始用户管理模块全面测试...\n');
    
    const tests = [
        { name: '管理员登录', fn: testAdminLogin },
        { name: '创建用户', fn: testCreateUser },
        { name: '获取用户列表', fn: testGetUserList },
        { name: '根据ID获取用户', fn: testGetUserById },
        { name: '更新用户信息', fn: testUpdateUser },
        { name: '搜索用户', fn: testSearchUsers },
        { name: '用户状态切换', fn: testUserStatusToggle },
        { name: '删除用户', fn: testDeleteUser },
        { name: '验证用户已删除', fn: testVerifyUserDeleted }
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
    
    console.log('\n📊 用户管理模块测试结果汇总:');
    console.log(`✅ 通过: ${passedTests}/${totalTests}`);
    console.log(`❌ 失败: ${totalTests - passedTests}/${totalTests}`);
    
    if (passedTests === totalTests) {
        console.log('🎉 所有用户管理测试通过！');
    } else {
        console.log('⚠️  部分用户管理测试失败，请检查相关功能。');
    }
}

// 运行测试
runUserManagementTests().catch(console.error);