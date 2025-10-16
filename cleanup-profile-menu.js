/**
 * 清理profile菜单及其子菜单
 */

async function cleanupProfileMenu() {
  const baseURL = 'http://localhost:8080';
  
  try {
    console.log('🧹 开始清理profile菜单...');
    
    // 1. 获取菜单树
    console.log('🌳 获取菜单树...');
    const treeResponse = await fetch(`${baseURL}/api/v1/menus/tree`, {
      headers: {
        'Authorization': 'Bearer test-token'
      }
    });
    
    if (!treeResponse.ok) {
      throw new Error('获取菜单树失败');
    }
    
    const menuTree = await treeResponse.json();
    const menus = menuTree.data || [];
    
    // 2. 找到profile菜单
    const profileMenu = menus.find(menu => menu.name === 'profile');
    
    if (!profileMenu) {
      console.log('✅ profile菜单已不存在');
      return;
    }
    
    console.log(`🔍 找到profile菜单 (ID: ${profileMenu.id})`);
    console.log(`   └─ 包含 ${profileMenu.children?.length || 0} 个子菜单`);
    
    // 3. 先删除所有子菜单
    if (profileMenu.children && profileMenu.children.length > 0) {
      console.log('🗑️ 删除子菜单...');
      
      for (const child of profileMenu.children) {
        try {
          const deleteResponse = await fetch(`${baseURL}/api/v1/admin/menus/${child.id}`, {
            method: 'DELETE',
            headers: {
              'Authorization': 'Bearer test-token'
            }
          });
          
          if (deleteResponse.ok) {
            console.log(`✅ 删除子菜单: ${child.name}`);
          } else {
            const errorText = await deleteResponse.text();
            console.log(`❌ 删除子菜单失败: ${child.name} - ${errorText}`);
          }
        } catch (error) {
          console.log(`❌ 删除子菜单异常: ${child.name} - ${error.message}`);
        }
      }
    }
    
    // 4. 删除profile主菜单
    console.log('🗑️ 删除profile主菜单...');
    try {
      const deleteResponse = await fetch(`${baseURL}/api/v1/admin/menus/${profileMenu.id}`, {
        method: 'DELETE',
        headers: {
          'Authorization': 'Bearer test-token'
        }
      });
      
      if (deleteResponse.ok) {
        console.log(`✅ 删除主菜单: ${profileMenu.name}`);
      } else {
        const errorText = await deleteResponse.text();
        console.log(`❌ 删除主菜单失败: ${profileMenu.name} - ${errorText}`);
      }
    } catch (error) {
      console.log(`❌ 删除主菜单异常: ${profileMenu.name} - ${error.message}`);
    }
    
    // 5. 验证清理结果
    console.log('🔍 验证清理结果...');
    const finalResponse = await fetch(`${baseURL}/api/v1/menus/tree`, {
      headers: {
        'Authorization': 'Bearer test-token'
      }
    });
    
    if (finalResponse.ok) {
      const finalMenuTree = await finalResponse.json();
      console.log('✅ 菜单树获取成功');
      console.log('📊 清理后菜单统计:');
      console.log(`   - 主菜单数量: ${finalMenuTree.data?.length || 0}`);
      
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
      
      totalMenus = countMenus(finalMenuTree.data);
      console.log(`   - 总菜单数量: ${totalMenus}`);
      
      // 检查profile菜单是否还存在
      const stillHasProfile = finalMenuTree.data?.some(menu => menu.name === 'profile');
      
      if (stillHasProfile) {
        console.log('\n⚠️ profile菜单仍然存在');
      } else {
        console.log('\n✅ profile菜单已成功删除');
      }
      
    } else {
      console.log('❌ 最终验证失败');
    }
    
    console.log('\n🎉 profile菜单清理完成！');
    
  } catch (error) {
    console.error('❌ profile菜单清理失败:', error.message);
  }
}

// 运行清理
cleanupProfileMenu();