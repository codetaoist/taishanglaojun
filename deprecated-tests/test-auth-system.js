// 使用Node.js内置的fetch API (Node.js 18+)

// 全局变量
let adminToken = '';
let userToken = '';
let testUserId = '';

// 基础URL
const baseURL = 'http://localhost:8080';

// 1. 测试管理员登录
async function testAdminLogin() {
    try {
        console.log('1️⃣ 测试管理员登录...');
        const response = await fetch(`${baseURL}/api/v1/auth/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                username: 'admin',
                password: 'admin123'
            })
        });

        if (!response.ok) {
            throw new Error(`管理员登录失败: ${response.status} ${response.statusText}`);
        }

        const data = await response.json();
        adminToken = data.data.token;
        
        console.log('✅ 管理员登录成功');
        console.log(`   - 用户ID: ${data.data.user_id}`);
        console.log(`   - 用户名: ${data.data.username}`);
        console.log(`   - 邮箱: ${data.data.email}`);
        console.log(`   - Token长度: ${adminToken.length}`);
        
        return true;
    } catch (error) {
        console.error('❌ 管理员登录失败:', error.message);
        return false;
    }
}

// 2. 测试错误登录凭据
async function testInvalidLogin() {
    try {
        console.log('2️⃣ 测试错误登录凭据...');
        const response = await fetch(`${baseURL}/api/v1/auth/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                username: 'admin',
                password: 'wrongpassword'
            })
        });

        if (response.ok) {
            console.log('❌ 错误凭据登录应该失败但却成功了');
            return false;
        } else {
            console.log('✅ 错误凭据登录正确被拒绝');
            console.log(`   - 状态码: ${response.status}`);
            return true;
        }
    } catch (error) {
        console.error('❌ 测试错误登录失败:', error.message);
        return false;
    }
}

// 3. 测试获取当前用户信息
async function testGetCurrentUser() {
    try {
        console.log('3️⃣ 测试获取当前用户信息...');
        const response = await fetch(`${baseURL}/api/v1/auth/me`, {
            headers: {
                'Authorization': `Bearer ${adminToken}`
            }
        });

        if (!response.ok) {
            throw new Error(`获取用户信息失败: ${response.status} ${response.statusText}`);
        }

        const data = await response.json();
        console.log('✅ 获取当前用户信息成功');
        console.log(`   - 用户名: ${data.data.username}`);
        console.log(`   - 角色: ${data.data.role}`);
        console.log(`   - 等级: ${data.data.level}`);
        
        return true;
    } catch (error) {
        console.error('❌ 获取用户信息失败:', error.message);
        return false;
    }
}

// 4. 测试无效token访问
async function testInvalidTokenAccess() {
    try {
        console.log('4️⃣ 测试无效token访问...');
        const response = await fetch(`${baseURL}/api/v1/auth/me`, {
            headers: {
                'Authorization': 'Bearer invalid_token_here'
            }
        });

        if (response.ok) {
            console.log('❌ 无效token访问应该失败但却成功了');
            return false;
        } else {
            console.log('✅ 无效token访问正确被拒绝');
            console.log(`   - 状态码: ${response.status}`);
            return true;
        }
    } catch (error) {
        console.error('❌ 测试无效token访问失败:', error.message);
        return false;
    }
}

// 5. 测试用户注册
async function testUserRegistration() {
    try {
        console.log('5️⃣ 测试用户注册...');
        const userData = {
            username: `testuser_${Date.now()}`,
            password: 'test123456',
            email: `test_${Date.now()}@example.com`
        };

        const response = await fetch(`${baseURL}/api/v1/auth/register`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(userData)
        });

        if (!response.ok) {
            throw new Error(`用户注册失败: ${response.status} ${response.statusText}`);
        }

        const data = await response.json();
        testUserId = data.data.user_id;
        
        console.log('✅ 用户注册成功');
        console.log(`   - 用户ID: ${data.data.user_id}`);
        console.log(`   - 用户名: ${data.data.username}`);
        console.log(`   - 邮箱: ${data.data.email}`);
        
        // 尝试用新注册的用户登录
        const loginResponse = await fetch(`${baseURL}/api/v1/auth/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                username: userData.username,
                password: userData.password
            })
        });

        if (loginResponse.ok) {
            const loginData = await loginResponse.json();
            userToken = loginData.data.token;
            console.log('✅ 新注册用户登录成功');
            return true;
        } else {
            console.log('❌ 新注册用户登录失败');
            return false;
        }
    } catch (error) {
        console.error('❌ 用户注册失败:', error.message);
        return false;
    }
}

// 6. 测试重复用户名注册
async function testDuplicateUsernameRegistration() {
    try {
        console.log('6️⃣ 测试重复用户名注册...');
        const response = await fetch(`${baseURL}/api/v1/auth/register`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                username: 'admin', // 使用已存在的用户名
                password: 'test123456',
                email: 'duplicate@example.com'
            })
        });

        if (response.ok) {
            console.log('❌ 重复用户名注册应该失败但却成功了');
            return false;
        } else {
            console.log('✅ 重复用户名注册正确被拒绝');
            console.log(`   - 状态码: ${response.status}`);
            return true;
        }
    } catch (error) {
        console.error('❌ 测试重复用户名注册失败:', error.message);
        return false;
    }
}

// 7. 测试权限验证
async function testPermissionValidation() {
    try {
        console.log('7️⃣ 测试权限验证...');
        
        // 管理员访问管理员接口
        const adminResponse = await fetch(`${baseURL}/api/v1/admin/users`, {
            headers: {
                'Authorization': `Bearer ${adminToken}`
            }
        });

        // 普通用户访问管理员接口
        const userResponse = await fetch(`${baseURL}/api/v1/admin/users`, {
            headers: {
                'Authorization': `Bearer ${userToken}`
            }
        });

        const adminCanAccess = adminResponse.ok;
        const userCanAccess = userResponse.ok;

        console.log(`   - 管理员访问管理员接口: ${adminCanAccess ? '成功' : '失败'}`);
        console.log(`   - 普通用户访问管理员接口: ${userCanAccess ? '成功' : '失败'}`);

        // 期望管理员可以访问，普通用户不能访问
        if (adminCanAccess && !userCanAccess) {
            console.log('✅ 权限验证正常工作');
            return true;
        } else {
            console.log('❌ 权限验证存在问题');
            return false;
        }
    } catch (error) {
        console.error('❌ 权限验证测试失败:', error.message);
        return false;
    }
}

// 8. 测试Token过期处理（模拟）
async function testTokenExpiration() {
    try {
        console.log('8️⃣ 测试Token过期处理...');
        
        // 使用一个明显过期的token
        const expiredToken = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMSIsInVzZXJuYW1lIjoidGVzdCIsImV4cCI6MTAwMDAwMDAwMH0.invalid';
        
        const response = await fetch(`${baseURL}/api/v1/auth/me`, {
            headers: {
                'Authorization': `Bearer ${expiredToken}`
            }
        });

        if (response.ok) {
            console.log('❌ 过期token访问应该失败但却成功了');
            return false;
        } else {
            console.log('✅ 过期token访问正确被拒绝');
            console.log(`   - 状态码: ${response.status}`);
            return true;
        }
    } catch (error) {
        console.error('❌ 测试Token过期处理失败:', error.message);
        return false;
    }
}

// 9. 测试创建测试用户（管理员功能）
async function testCreateTestUser() {
    try {
        console.log('9️⃣ 测试创建测试用户（管理员功能）...');
        const response = await fetch(`${baseURL}/api/v1/auth/test-user`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${adminToken}`
            }
        });

        if (!response.ok) {
            throw new Error(`创建测试用户失败: ${response.status} ${response.statusText}`);
        }

        const data = await response.json();
        console.log('✅ 创建测试用户成功');
        console.log(`   - 用户名: ${data.data.username}`);
        console.log(`   - 密码: ${data.data.password}`);
        
        return true;
    } catch (error) {
        console.error('❌ 创建测试用户失败:', error.message);
        return false;
    }
}

// 主测试函数
async function runAuthTests() {
    console.log('🔐 开始认证系统全面测试...\n');

    const tests = [
        { name: '管理员登录', func: testAdminLogin },
        { name: '错误登录凭据', func: testInvalidLogin },
        { name: '获取当前用户信息', func: testGetCurrentUser },
        { name: '无效token访问', func: testInvalidTokenAccess },
        { name: '用户注册', func: testUserRegistration },
        { name: '重复用户名注册', func: testDuplicateUsernameRegistration },
        { name: '权限验证', func: testPermissionValidation },
        { name: 'Token过期处理', func: testTokenExpiration },
        { name: '创建测试用户', func: testCreateTestUser }
    ];

    let passedTests = 0;
    let totalTests = tests.length;

    for (const test of tests) {
        console.log(`\n--- 测试: ${test.name} ---`);
        const result = await test.func();
        if (result) {
            passedTests++;
            console.log(`✅ ${test.name} - 通过`);
        } else {
            console.log(`❌ ${test.name} - 失败`);
        }
        
        // 在测试之间添加短暂延迟
        await new Promise(resolve => setTimeout(resolve, 500));
    }

    console.log('\n📊 认证系统测试结果汇总:');
    console.log(`✅ 通过: ${passedTests}/${totalTests}`);
    console.log(`❌ 失败: ${totalTests - passedTests}/${totalTests}`);
    
    if (passedTests === totalTests) {
        console.log('🎉 所有认证系统测试通过！');
    } else {
        console.log('⚠️  部分认证系统测试失败，请检查相关功能。');
    }

    return { passed: passedTests, total: totalTests };
}

// 运行测试
runAuthTests().catch(error => {
    console.error('认证系统测试运行出错:', error);
    process.exit(1);
});