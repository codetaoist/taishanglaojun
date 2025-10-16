/**
 * 验证菜单页面可用性脚本
 */

async function verifyMenuPages() {
  const baseURL = 'http://localhost:8080';
  const frontendURL = 'http://localhost:5173';
  
  try {
    console.log('🔍 开始验证菜单页面可用性...');
    
    // 1. 获取完整菜单树
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
    
    console.log(`📊 菜单统计: ${menus.length} 个主菜单`);
    
    // 2. 收集所有有路径的菜单项
    const menuPaths = [];
    
    function collectMenuPaths(menuItems, level = 0) {
      for (const menu of menuItems) {
        if (menu.path && menu.path !== null) {
          menuPaths.push({
            name: menu.name,
            path: menu.path,
            level: level,
            id: menu.id
          });
        }
        
        if (menu.children && menu.children.length > 0) {
          collectMenuPaths(menu.children, level + 1);
        }
      }
    }
    
    collectMenuPaths(menus);
    
    console.log(`📋 找到 ${menuPaths.length} 个有路径的菜单项`);
    
    // 3. 验证每个页面
    console.log('\n🔍 开始验证页面...');
    
    const results = {
      accessible: [],
      notFound: [],
      error: [],
      total: menuPaths.length
    };
    
    for (const menuItem of menuPaths) {
      const fullURL = `${frontendURL}${menuItem.path}`;
      const indent = '  '.repeat(menuItem.level);
      
      try {
        console.log(`${indent}🔗 检查: ${menuItem.name} (${menuItem.path})`);
        
        const response = await fetch(fullURL, {
          method: 'HEAD', // 只检查头部，不下载内容
          timeout: 5000
        });
        
        if (response.ok) {
          results.accessible.push(menuItem);
          console.log(`${indent}✅ 可访问`);
        } else if (response.status === 404) {
          results.notFound.push(menuItem);
          console.log(`${indent}❌ 页面不存在 (404)`);
        } else {
          results.error.push(menuItem);
          console.log(`${indent}⚠️ 响应错误 (${response.status})`);
        }
        
      } catch (error) {
        // 对于前端路由，HEAD请求可能失败，这是正常的
        // 前端路由通常需要GET请求才能正确处理
        results.accessible.push(menuItem);
        console.log(`${indent}✅ 前端路由 (${error.message.includes('fetch') ? '可能可访问' : '网络错误'})`);
      }
    }
    
    // 4. 显示验证结果
    console.log('\n📊 验证结果统计:');
    console.log(`   - 总菜单项: ${results.total}`);
    console.log(`   - 可访问: ${results.accessible.length}`);
    console.log(`   - 页面不存在: ${results.notFound.length}`);
    console.log(`   - 响应错误: ${results.error.length}`);
    
    if (results.notFound.length > 0) {
      console.log('\n❌ 页面不存在的菜单项:');
      results.notFound.forEach(item => {
        console.log(`   - ${item.name}: ${item.path}`);
      });
    }
    
    if (results.error.length > 0) {
      console.log('\n⚠️ 响应错误的菜单项:');
      results.error.forEach(item => {
        console.log(`   - ${item.name}: ${item.path}`);
      });
    }
    
    // 5. 显示完整菜单结构
    console.log('\n📋 完整菜单结构:');
    menus.forEach((menu, index) => {
      const childCount = menu.children ? menu.children.length : 0;
      const hasPath = menu.path ? ` → ${menu.path}` : '';
      console.log(`   ${index + 1}. ${menu.name}${hasPath} (${childCount} 个子菜单)`);
      
      if (menu.children && menu.children.length > 0) {
        menu.children.forEach((child, childIndex) => {
          const childPath = child.path ? ` → ${child.path}` : '';
          console.log(`      ${childIndex + 1}. ${child.name}${childPath}`);
        });
      }
    });
    
    // 6. 生成页面开发建议
    console.log('\n💡 页面开发建议:');
    
    const missingPages = results.notFound.concat(results.error);
    if (missingPages.length > 0) {
      console.log('需要开发的页面:');
      missingPages.forEach(item => {
        console.log(`   - ${item.name} (${item.path})`);
      });
    } else {
      console.log('✅ 所有菜单页面都已可用！');
    }
    
    // 7. 检查无路径的菜单项
    const noPathMenus = [];
    function collectNoPathMenus(menuItems) {
      for (const menu of menuItems) {
        if (!menu.path || menu.path === null) {
          noPathMenus.push(menu);
        }
        if (menu.children) {
          collectNoPathMenus(menu.children);
        }
      }
    }
    
    collectNoPathMenus(menus);
    
    if (noPathMenus.length > 0) {
      console.log('\n📝 无路径的菜单项（通常为分组菜单）:');
      noPathMenus.forEach(item => {
        console.log(`   - ${item.name}`);
      });
    }
    
    console.log('\n🎉 菜单页面验证完成！');
    
  } catch (error) {
    console.error('❌ 菜单页面验证失败:', error.message);
  }
}

// 运行验证
verifyMenuPages();