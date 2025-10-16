/**
 * 清理重复菜单项脚本 - 通过菜单树识别重复项
 */

async function cleanupDuplicateMenus() {
  const baseURL = 'http://localhost:8080';
  
  try {
    console.log('🧹 开始清理重复菜单项...');
    
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
    
    console.log(`📊 当前主菜单数量: ${menus.length}`);
    
    // 2. 识别重复菜单
    const duplicateMenus = [];
    
    // 需要删除的重复菜单（保留中文版本）
    const duplicatePatterns = [
      { keep: '仪表板', remove: ['dashboard'] },
      { keep: '个人中心', remove: ['profile'] },
      { keep: 'AI智能服务', remove: ['ai-chat'] },
      { keep: '帮助中心', remove: ['help'] }
    ];
    
    for (const menu of menus) {
      for (const pattern of duplicatePatterns) {
        if (pattern.remove.includes(menu.name)) {
          duplicateMenus.push(menu);
          console.log(`🔍 发现重复菜单: ${menu.name} (ID: ${menu.id})`);
          
          // 如果有子菜单，也需要删除
          if (menu.children && menu.children.length > 0) {
            console.log(`   └─ 包含 ${menu.children.length} 个子菜单`);
          }
        }
      }
    }
    
    // 3. 删除重复菜单
    console.log(`🗑️ 准备删除 ${duplicateMenus.length} 个重复菜单...`);
    
    for (const menu of duplicateMenus) {
      try {
        const deleteResponse = await fetch(`${baseURL}/api/v1/admin/menus/${menu.id}`, {
          method: 'DELETE',
          headers: {
            'Authorization': 'Bearer test-token'
          }
        });
        
        if (deleteResponse.ok) {
          console.log(`✅ 删除重复菜单: ${menu.name}`);
        } else {
          const errorText = await deleteResponse.text();
          console.log(`❌ 删除菜单失败: ${menu.name} - ${errorText}`);
        }
      } catch (error) {
        console.log(`❌ 删除菜单异常: ${menu.name} - ${error.message}`);
      }
    }
    
    // 4. 验证清理结果
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
      
      // 显示最终菜单结构
      console.log('\n📋 清理后菜单结构:');
      finalMenuTree.data?.forEach((menu, index) => {
        const childCount = menu.children ? menu.children.length : 0;
        console.log(`   ${index + 1}. ${menu.name} (${childCount} 个子菜单)`);
        if (menu.children && menu.children.length > 0) {
          menu.children.forEach((child, childIndex) => {
            console.log(`      ${childIndex + 1}. ${child.name}`);
          });
        }
      });
      
      // 检查是否还有重复项
      const finalMenuNames = finalMenuTree.data?.map(m => m.name) || [];
      const duplicateCheck = duplicatePatterns.some(pattern => 
        pattern.remove.some(name => finalMenuNames.includes(name))
      );
      
      if (duplicateCheck) {
        console.log('\n⚠️ 仍有重复菜单项存在');
      } else {
        console.log('\n✅ 所有重复菜单项已清理完成');
      }
      
    } else {
      console.log('❌ 最终验证失败');
    }
    
    console.log('\n🎉 菜单清理完成！');
    
  } catch (error) {
    console.error('❌ 菜单清理失败:', error.message);
  }
}

// 运行清理
cleanupDuplicateMenus();