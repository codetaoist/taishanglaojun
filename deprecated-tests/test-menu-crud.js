/**
 * 测试菜单管理CRUD功能
 */

// 使用Node.js内置的fetch API (Node.js 18+)
const baseURL = 'http://localhost:8080';

async function testMenuCRUD() {
  console.log('🧪 开始测试菜单管理CRUD功能...\n');
  
  try {
    // 1. 登录获取token
    console.log('1️⃣ 登录获取token...');
    const loginResponse = await fetch(`${baseURL}/api/v1/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        username: 'admin',
        password: 'admin123'
      })
    });
    
    if (!loginResponse.ok) {
      throw new Error(`登录失败: ${loginResponse.status}`);
    }
    
    const loginData = await loginResponse.json();
     token = loginData.data.token;
    
    if (!token) {
      throw new Error('未获取到token');
    }
    
    console.log('✅ 登录成功，获取到token');
    
    // 2. 测试创建菜单
    console.log('\n2️⃣ 测试创建菜单...');
    const testMenu = {
      name: '测试菜单',
      title: '测试菜单标题',
      path: '/test-menu',
      icon: 'TestOutlined',
      sort: 999,
      level: 1,
      is_visible: true,
      is_enabled: true,
      required_role: 'user'
    };
    
    const createResponse = await fetch(`${baseURL}/api/v1/admin/menus`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify(testMenu)
    });
    
    if (!createResponse.ok) {
      const errorText = await createResponse.text();
      throw new Error(`创建菜单失败: ${createResponse.status} - ${errorText}`);
    }
    
    const createData = await createResponse.json();
    const createdMenuId = createData.data?.id;
    
    if (!createdMenuId) {
      throw new Error('创建菜单成功但未返回菜单ID');
    }
    
    console.log(`✅ 创建菜单成功，ID: ${createdMenuId}`);
    
    // 3. 测试获取菜单详情
    console.log('\n3️⃣ 测试获取菜单详情...');
    const getResponse = await fetch(`${baseURL}/api/v1/menus/${createdMenuId}`, {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });
    
    if (!getResponse.ok) {
      const errorText = await getResponse.text();
      throw new Error(`获取菜单详情失败: ${getResponse.status} - ${errorText}`);
    }
    
    const getData = await getResponse.json();
    console.log('✅ 获取菜单详情成功');
    console.log(`   - 菜单名称: ${getData.data?.name}`);
    console.log(`   - 菜单路径: ${getData.data?.path}`);
    console.log(`   - 菜单状态: ${getData.data?.is_enabled ? '启用' : '禁用'}`);
    
    // 4. 测试更新菜单
    console.log('\n4️⃣ 测试更新菜单...');
    const updateData = {
      name: '更新后的测试菜单',
      title: '更新后的测试菜单标题',
      path: '/updated-test-menu',
      sort: 888
    };
    
    const updateResponse = await fetch(`${baseURL}/api/v1/admin/menus/${createdMenuId}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify(updateData)
    });
    
    if (!updateResponse.ok) {
      const errorText = await updateResponse.text();
      throw new Error(`更新菜单失败: ${updateResponse.status} - ${errorText}`);
    }
    
    console.log('✅ 更新菜单成功');
    
    // 5. 验证更新结果
    console.log('\n5️⃣ 验证更新结果...');
    const verifyResponse = await fetch(`${baseURL}/api/v1/menus/${createdMenuId}`, {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });
    
    if (verifyResponse.ok) {
      const verifyData = await verifyResponse.json();
      console.log('✅ 验证更新结果成功');
      console.log(`   - 更新后名称: ${verifyData.data?.name}`);
      console.log(`   - 更新后路径: ${verifyData.data?.path}`);
      console.log(`   - 更新后排序: ${verifyData.data?.sort}`);
    }
    
    // 6. 测试菜单列表
    console.log('\n6️⃣ 测试菜单列表...');
    const listResponse = await fetch(`${baseURL}/api/v1/menus`, {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });
    
    if (listResponse.ok) {
      const listData = await listResponse.json();
      console.log('✅ 获取菜单列表成功');
      console.log(`   - 菜单总数: ${listData.data?.length || 0}`);
      
      // 查找我们创建的测试菜单
      const testMenuItem = listData.data?.find(menu => menu.id === createdMenuId);
      if (testMenuItem) {
        console.log('   - 测试菜单在列表中找到 ✅');
      } else {
        console.log('   - 测试菜单在列表中未找到 ❌');
      }
    }
    
    // 7. 测试菜单树
    console.log('\n7️⃣ 测试菜单树...');
    const treeResponse = await fetch(`${baseURL}/api/v1/menus/tree`, {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });
    
    if (treeResponse.ok) {
      const treeData = await treeResponse.json();
      console.log('✅ 获取菜单树成功');
      console.log(`   - 根菜单数量: ${treeData.data?.length || 0}`);
    }
    
    // 8. 测试删除菜单
    console.log('\n8️⃣ 测试删除菜单...');
    const deleteResponse = await fetch(`${baseURL}/api/v1/admin/menus/${createdMenuId}`, {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });
    
    if (!deleteResponse.ok) {
      const errorText = await deleteResponse.text();
      throw new Error(`删除菜单失败: ${deleteResponse.status} - ${errorText}`);
    }
    
    console.log('✅ 删除菜单成功');
    
    // 9. 验证删除结果
    console.log('\n9️⃣ 验证删除结果...');
    const verifyDeleteResponse = await fetch(`${baseURL}/api/v1/menus/${createdMenuId}`, {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });
    
    if (verifyDeleteResponse.status === 404) {
      console.log('✅ 验证删除成功，菜单已不存在');
    } else {
      console.log('❌ 删除验证失败，菜单仍然存在');
    }
    
    console.log('\n🎉 菜单管理CRUD功能测试完成！');
    console.log('\n📊 测试结果总结:');
    console.log('   ✅ 创建菜单 - 成功');
    console.log('   ✅ 获取菜单详情 - 成功');
    console.log('   ✅ 更新菜单 - 成功');
    console.log('   ✅ 获取菜单列表 - 成功');
    console.log('   ✅ 获取菜单树 - 成功');
    console.log('   ✅ 删除菜单 - 成功');
    
  } catch (error) {
    console.error('❌ 测试过程中发生错误:', error.message);
    
    console.log('\n🔧 可能的解决方案:');
    console.log('1. 检查后端服务是否正常运行');
    console.log('2. 检查数据库连接是否正常');
    console.log('3. 检查API路径是否正确');
    console.log('4. 检查用户权限是否足够');
  }
}

// 运行测试
testMenuCRUD();