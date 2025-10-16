// 使用Node.js内置的fetch API (Node.js 18+)

// 全局变量
let adminToken = '';
let userToken = '';
let testMenuId = '';

// 基础URL
const baseURL = 'http://localhost:8080';

// 1. 管理员登录
async function adminLogin() {
    try {
        console.log('1️⃣ 管理员登录...');
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
        return true;
    } catch (error) {
        console.error('❌ 管理员登录失败:', error.message);
        return false;
    }
}

// 2. 创建测试菜单（需要特定权限）
async function createTestMenu() {
    try {
        console.log('2️⃣ 创建需要特定权限的测试菜单...');
        const menuData = {
            name: '权限测试菜单',
            path: '/permission-test',
            icon: 'PermissionIcon',
            description: '这是一个需要特定权限的菜单',
            status: 'active',
            parent_id: null,
            sort_order: 999,
            required_permission: 'admin:permission:test'
        };

        const response = await fetch(`${baseURL}/api/v1/admin/menus`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${adminToken}`
            },
            body: JSON.stringify(menuData)
        });

        if (!response.ok) {
            throw new Error(`创建测试菜单失败: ${response.status} ${response.statusText}`);
        }

        const data = await response.json();
        testMenuId = data.id;
        console.log('✅ 权限测试菜单创建成功, ID:', testMenuId);
        return true;
    } catch (error) {
        console.error('❌ 创建测试菜单失败:', error.message);
        return false;
    }
}

// 3. 管理员获取菜单树（应该包含所有菜单）
async function adminGetMenuTree() {
    try {
        console.log('3️⃣ 管理员获取菜单树...');
        const response = await fetch(`${baseURL}/api/v1/menus/tree`, {
            headers: {
                'Authorization': `Bearer ${adminToken}`
            }
        });

        if (!response.ok) {
            throw new Error(`管理员获取菜单树失败: ${response.status} ${response.statusText}`);
        }

        const data = await response.json();
        
        // 检查是否包含测试菜单
        const hasTestMenu = JSON.stringify(data).includes('权限测试菜单');
        
        console.log('✅ 管理员获取菜单树成功');
        console.log(`   - 根菜单数量: ${data.length}`);
        console.log(`   - 包含权限测试菜单: ${hasTestMenu ? '是' : '否'}`);
        
        return hasTestMenu;
    } catch (error) {
        console.error('❌ 管理员获取菜单树失败:', error.message);
        return false;
    }
}

// 4. 尝试创建普通用户（如果不存在）
async function createTestUser() {
    try {
        console.log('4️⃣ 创建测试用户...');
        const userData = {
            username: 'testuser',
            password: 'test123',
            email: 'test@example.com',
            role: 'user'
        };

        const response = await fetch(`${baseURL}/api/v1/admin/users`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${adminToken}`
            },
            body: JSON.stringify(userData)
        });

        if (response.ok) {
            console.log('✅ 测试用户创建成功');
            return true;
        } else if (response.status === 409) {
            console.log('ℹ️  测试用户已存在');
            return true;
        } else {
            throw new Error(`创建测试用户失败: ${response.status} ${response.statusText}`);
        }
    } catch (error) {
        console.error('❌ 创建测试用户失败:', error.message);
        return false;
    }
}

// 5. 普通用户登录
async function userLogin() {
    try {
        console.log('5️⃣ 普通用户登录...');
        const response = await fetch(`${baseURL}/api/v1/auth/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                username: 'testuser',
                password: 'test123'
            })
        });

        if (!response.ok) {
            throw new Error(`普通用户登录失败: ${response.status} ${response.statusText}`);
        }

        const data = await response.json();
        userToken = data.data.token;
        console.log('✅ 普通用户登录成功');
        return true;
    } catch (error) {
        console.error('❌ 普通用户登录失败:', error.message);
        return false;
    }
}

// 6. 普通用户获取菜单树（应该过滤掉没有权限的菜单）
async function userGetMenuTree() {
    try {
        console.log('6️⃣ 普通用户获取菜单树...');
        const response = await fetch(`${baseURL}/api/v1/menus/tree`, {
            headers: {
                'Authorization': `Bearer ${userToken}`
            }
        });

        if (!response.ok) {
            throw new Error(`普通用户获取菜单树失败: ${response.status} ${response.statusText}`);
        }

        const data = await response.json();
        
        // 检查是否包含测试菜单（应该不包含）
        const hasTestMenu = JSON.stringify(data).includes('权限测试菜单');
        
        console.log('✅ 普通用户获取菜单树成功');
        console.log(`   - 根菜单数量: ${data.length}`);
        console.log(`   - 包含权限测试菜单: ${hasTestMenu ? '是' : '否'}`);
        
        // 权限过滤正常工作的话，普通用户不应该看到权限测试菜单
        return !hasTestMenu;
    } catch (error) {
        console.error('❌ 普通用户获取菜单树失败:', error.message);
        return false;
    }
}

// 7. 测试权限检查API
async function testPermissionCheck() {
    try {
        console.log('7️⃣ 测试权限检查API...');
        
        // 管理员检查权限
        const adminResponse = await fetch(`${baseURL}/api/v1/permissions/check`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${adminToken}`
            },
            body: JSON.stringify({
                permission: 'admin:permission:test'
            })
        });

        // 普通用户检查权限
        const userResponse = await fetch(`${baseURL}/api/v1/permissions/check`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${userToken}`
            },
            body: JSON.stringify({
                permission: 'admin:permission:test'
            })
        });

        const adminHasPermission = adminResponse.ok && (await adminResponse.json()).hasPermission;
        const userHasPermission = userResponse.ok && (await userResponse.json()).hasPermission;

        console.log(`   - 管理员有权限: ${adminHasPermission ? '是' : '否'}`);
        console.log(`   - 普通用户有权限: ${userHasPermission ? '是' : '否'}`);

        // 期望管理员有权限，普通用户没有权限
        return adminHasPermission && !userHasPermission;
    } catch (error) {
        console.error('❌ 权限检查测试失败:', error.message);
        return false;
    }
}

// 8. 清理测试数据
async function cleanup() {
    try {
        console.log('8️⃣ 清理测试数据...');
        
        // 删除测试菜单
        if (testMenuId) {
            const deleteMenuResponse = await fetch(`${baseURL}/api/v1/admin/menus/${testMenuId}`, {
                method: 'DELETE',
                headers: {
                    'Authorization': `Bearer ${adminToken}`
                }
            });
            
            if (deleteMenuResponse.ok) {
                console.log('✅ 测试菜单删除成功');
            }
        }

        return true;
    } catch (error) {
        console.error('❌ 清理测试数据失败:', error.message);
        return false;
    }
}

// 主测试函数
async function runPermissionTests() {
    console.log('🔐 开始菜单权限过滤功能测试...\n');

    const tests = [
        { name: '管理员登录', func: adminLogin },
        { name: '创建权限测试菜单', func: createTestMenu },
        { name: '管理员获取菜单树', func: adminGetMenuTree },
        { name: '创建测试用户', func: createTestUser },
        { name: '普通用户登录', func: userLogin },
        { name: '普通用户获取菜单树', func: userGetMenuTree },
        { name: '权限检查API测试', func: testPermissionCheck },
        { name: '清理测试数据', func: cleanup }
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

    console.log('\n📊 权限测试结果汇总:');
    console.log(`✅ 通过: ${passedTests}/${totalTests}`);
    console.log(`❌ 失败: ${totalTests - passedTests}/${totalTests}`);
    
    if (passedTests === totalTests) {
        console.log('🎉 所有权限测试通过！菜单权限过滤功能正常工作。');
    } else {
        console.log('⚠️  部分权限测试失败，请检查相关功能。');
    }
}

// 运行测试
runPermissionTests().catch(error => {
    console.error('权限测试运行出错:', error);
    process.exit(1);
});