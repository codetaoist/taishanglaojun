// 调试菜单权限过滤问题
console.log('🔍 开始调试菜单权限过滤...');

// 1. 检查用户信息
async function checkUserInfo() {
  try {
    const response = await fetch('http://localhost:8080/api/v1/auth/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        username: 'admin',
        password: 'admin123'
      })
    });
    
    const loginData = await response.json();
    console.log('🔐 登录响应:', loginData);
    
    if (loginData.success && loginData.data.token) {
      const userResponse = await fetch('http://localhost:8080/api/v1/auth/me', {
        headers: {
          'Authorization': `Bearer ${loginData.data.token}`
        }
      });
      
      const userData = await userResponse.json();
      console.log('👤 用户信息:', userData);
      
      if (userData.success) {
        const user = userData.data;
        console.log('🔍 用户详细信息:', {
          id: user.id,
          username: user.username,
          role: user.role,
          roles: user.roles,
          permissions: user.permissions,
          isAdmin: user.isAdmin
        });
        
        // 2. 检查菜单数据
        const menuResponse = await fetch('http://localhost:8080/api/v1/menu/tree', {
          headers: {
            'Authorization': `Bearer ${loginData.data.token}`
          }
        });
        
        const menuData = await menuResponse.json();
        console.log('📋 菜单数据:', menuData);
        
        if (menuData.success) {
          console.log('🔍 菜单项分析:');
          menuData.data.forEach((menu, index) => {
            console.log(`菜单 ${index + 1}:`, {
              name: menu.name,
              title: menu.title,
              required_role: menu.required_role,
              is_visible: menu.is_visible,
              is_enabled: menu.is_enabled,
              children: menu.children?.length || 0
            });
            
            if (menu.children) {
              menu.children.forEach((child, childIndex) => {
                console.log(`  子菜单 ${childIndex + 1}:`, {
                  name: child.name,
                  title: child.title,
                  required_role: child.required_role,
                  is_visible: child.is_visible,
                  is_enabled: child.is_enabled
                });
              });
            }
          });
          
          // 3. 模拟权限过滤逻辑
          console.log('🧪 模拟权限过滤逻辑:');
          const userPermissions = {
            roles: user.roles || ['user'],
            permissions: user.permissions || []
          };
          
          console.log('🔑 用户权限对象:', userPermissions);
          
          // 检查管理员权限
          const isAdmin = userPermissions.roles?.some(role => 
            ['admin', 'administrator', 'super_admin'].includes(role.toLowerCase())
          );
          console.log('👑 是否为管理员:', isAdmin);
          
          // 检查每个菜单项
          menuData.data.forEach(menu => {
            console.log(`\n🔍 检查菜单: ${menu.name || menu.title}`);
            
            // 处理required_role字段
            let requiredRole = [];
            if (menu.required_role) {
              if (typeof menu.required_role === 'string') {
                requiredRole = menu.required_role.split(',').map(role => role.trim().toLowerCase()).filter(Boolean);
              } else if (Array.isArray(menu.required_role)) {
                requiredRole = menu.required_role.map(role => role.toLowerCase());
              }
            }
            
            console.log(`  📝 所需角色: ${JSON.stringify(requiredRole)}`);
            console.log(`  👤 用户角色: ${JSON.stringify(userPermissions.roles)}`);
            
            if (isAdmin) {
              console.log(`  ✅ 管理员权限，允许访问`);
            } else if (requiredRole.length > 0) {
              const hasRole = requiredRole.some(role => {
                const includes = userPermissions.roles?.some(userRole => 
                  userRole.toLowerCase() === role.toLowerCase()
                );
                console.log(`    🔍 检查角色 "${role}": ${includes}`);
                return includes;
              });
              console.log(`  🔑 角色检查结果: ${hasRole ? '✅ 通过' : '❌ 拒绝'}`);
            } else {
              console.log(`  ✅ 无角色要求，允许访问`);
            }
          });
        }
      }
    }
  } catch (error) {
    console.error('❌ 调试过程中出错:', error);
  }
}

// 4. 检查前端菜单服务
function checkFrontendMenuService() {
  console.log('\n🔍 检查前端菜单服务...');
  
  // 检查是否有菜单上下文
  const menuContext = window.React?.useContext ? 'React上下文可用' : 'React上下文不可用';
  console.log('⚛️ React状态:', menuContext);
  
  // 检查localStorage中的菜单数据
  try {
    const savedMenuItems = localStorage.getItem('menuItems');
    if (savedMenuItems) {
      const menuItems = JSON.parse(savedMenuItems);
      console.log('💾 localStorage中的菜单数据:', menuItems);
    } else {
      console.log('💾 localStorage中没有菜单数据');
    }
  } catch (error) {
    console.log('💾 读取localStorage菜单数据失败:', error);
  }
}

// 执行调试
checkUserInfo().then(() => {
  checkFrontendMenuService();
  console.log('🎯 调试完成！请查看上面的日志分析问题。');
});