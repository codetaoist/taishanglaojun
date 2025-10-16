/**
 * 修复菜单显示问题脚本
 * 1. 分析API重复调用的原因
 * 2. 检查菜单数据转换逻辑
 * 3. 提供修复建议
 */

async function fixMenuDisplayIssue() {
  const baseURL = 'http://localhost:8080';
  
  try {
    console.log('🔧 开始修复菜单显示问题...');
    
    // 1. 先登录获取token
    console.log('\n1️⃣ 登录获取token...');
    const loginResponse = await fetch(`${baseURL}/api/v1/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
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
    console.log('✅ 登录成功');
    const token = loginData.token;
    
    // 2. 获取菜单数据并分析结构
    console.log('\n2️⃣ 获取菜单数据并分析结构...');
    const menuResponse = await fetch(`${baseURL}/api/v1/menus/tree`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      }
    });
    
    if (!menuResponse.ok) {
      throw new Error(`获取菜单失败: ${menuResponse.status}`);
    }
    
    const menuData = await menuResponse.json();
    console.log('📋 原始菜单数据结构:');
    console.log('   - 响应格式:', Object.keys(menuData));
    console.log('   - 菜单数量:', menuData.data ? menuData.data.length : 0);
    
    if (!menuData.data || menuData.data.length === 0) {
      console.log('❌ 菜单数据为空，这是菜单不显示的主要原因');
      return;
    }
    
    // 3. 分析菜单数据结构
    console.log('\n3️⃣ 分析菜单数据结构...');
    const menuItems = menuData.data;
    
    console.log('📊 菜单数据分析:');
    menuItems.forEach((menu, index) => {
      console.log(`   ${index + 1}. ${menu.name || menu.title}`);
      console.log(`      - ID: ${menu.id}`);
      console.log(`      - 路径: ${menu.path || '无'}`);
      console.log(`      - 图标: ${menu.icon || '无'}`);
      console.log(`      - 可见: ${menu.is_visible !== false ? '是' : '否'}`);
      console.log(`      - 启用: ${menu.is_enabled !== false ? '是' : '否'}`);
      console.log(`      - 排序: ${menu.sort || 0}`);
      console.log(`      - 父级ID: ${menu.parent_id || '无'}`);
      console.log(`      - 子菜单: ${menu.children ? menu.children.length : 0} 个`);
    });
    
    // 4. 模拟前端数据转换逻辑
    console.log('\n4️⃣ 模拟前端数据转换逻辑...');
    
    // 模拟 frontendMenuService.convertToFrontendFormat 的逻辑
    const convertToFrontendFormat = (backendMenus) => {
      return backendMenus
        .filter(menu => menu.is_visible !== false && menu.is_enabled !== false)
        .sort((a, b) => (a.sort || 0) - (b.sort || 0))
        .map(menu => {
          const frontendMenu = {
            key: menu.id.toString(),
            label: menu.title || menu.name,
            path: menu.path,
            icon: menu.icon,
            status: 'completed', // 默认状态
            requiredRole: [menu.required_role?.toLowerCase() || 'user'],
            requiredPermission: [],
            children: undefined
          };
          
          if (menu.children && menu.children.length > 0) {
            frontendMenu.children = convertToFrontendFormat(menu.children);
          }
          
          return frontendMenu;
        });
    };
    
    const frontendMenus = convertToFrontendFormat(menuItems);
    console.log('🔄 转换后的前端菜单:');
    console.log(`   - 菜单数量: ${frontendMenus.length}`);
    
    if (frontendMenus.length === 0) {
      console.log('❌ 前端转换后菜单为空！');
      console.log('🔍 可能的原因:');
      console.log('   1. 所有菜单都被 is_visible=false 或 is_enabled=false 过滤掉了');
      console.log('   2. 菜单数据结构不符合前端期望');
      console.log('   3. 转换逻辑有问题');
      
      // 检查被过滤的菜单
      const filteredOutMenus = menuItems.filter(menu => 
        menu.is_visible === false || menu.is_enabled === false
      );
      
      if (filteredOutMenus.length > 0) {
        console.log(`\n🚫 被过滤掉的菜单 (${filteredOutMenus.length}个):`);
        filteredOutMenus.forEach((menu, index) => {
          console.log(`   ${index + 1}. ${menu.name || menu.title}`);
          console.log(`      - 可见: ${menu.is_visible !== false ? '是' : '否'}`);
          console.log(`      - 启用: ${menu.is_enabled !== false ? '是' : '否'}`);
        });
      }
    } else {
      console.log('✅ 前端菜单转换成功');
      frontendMenus.forEach((menu, index) => {
        console.log(`   ${index + 1}. ${menu.label} (${menu.path || '无路径'})`);
        if (menu.children && menu.children.length > 0) {
          menu.children.forEach((child, childIndex) => {
            console.log(`      ${childIndex + 1}. ${child.label} (${child.path || '无路径'})`);
          });
        }
      });
    }
    
    // 5. 检查权限过滤逻辑
    console.log('\n5️⃣ 检查权限过滤逻辑...');
    
    // 模拟用户权限
    const userPermissions = {
      roles: ['admin'],
      permissions: ['admin:read', 'admin:write', 'user:read']
    };
    
    console.log('👤 模拟用户权限:', userPermissions);
    
    // 模拟权限过滤
    const filterByPermissions = (menus, permissions) => {
      return menus.filter(menu => {
        // 检查角色权限
        if (menu.requiredRole && menu.requiredRole.length > 0) {
          const hasRole = menu.requiredRole.some(role => 
            permissions.roles.includes(role)
          );
          if (!hasRole) {
            console.log(`   ❌ 菜单 "${menu.label}" 因角色权限被过滤 (需要: ${menu.requiredRole.join(', ')}, 拥有: ${permissions.roles.join(', ')})`);
            return false;
          }
        }
        
        // 检查具体权限
        if (menu.requiredPermission && menu.requiredPermission.length > 0) {
          const hasPermission = menu.requiredPermission.some(perm => 
            permissions.permissions.includes(perm)
          );
          if (!hasPermission) {
            console.log(`   ❌ 菜单 "${menu.label}" 因具体权限被过滤 (需要: ${menu.requiredPermission.join(', ')}, 拥有: ${permissions.permissions.join(', ')})`);
            return false;
          }
        }
        
        console.log(`   ✅ 菜单 "${menu.label}" 通过权限检查`);
        return true;
      });
    };
    
    const permissionFilteredMenus = filterByPermissions(frontendMenus, userPermissions);
    console.log(`🔒 权限过滤后的菜单数量: ${permissionFilteredMenus.length}`);
    
    // 6. 分析问题并提供解决方案
    console.log('\n6️⃣ 问题分析和解决方案...');
    
    const issues = [];
    const solutions = [];
    
    // 检查API重复调用问题
    issues.push('🔄 API被重复调用3次');
    solutions.push({
      problem: 'API重复调用',
      solution: '实现菜单数据的全局状态管理，避免多个组件重复请求',
      implementation: [
        '1. 创建MenuContext来管理菜单状态',
        '2. 在应用顶层提供菜单数据',
        '3. 各组件通过Context获取菜单数据，而不是独立请求',
        '4. 实现缓存机制，避免短时间内重复请求'
      ]
    });
    
    // 检查菜单显示问题
    if (frontendMenus.length === 0) {
      issues.push('📭 前端菜单数据为空');
      solutions.push({
        problem: '菜单不显示',
        solution: '修复菜单数据转换和过滤逻辑',
        implementation: [
          '1. 检查后端菜单数据的is_visible和is_enabled字段',
          '2. 修复前端数据转换逻辑',
          '3. 调整权限过滤条件',
          '4. 添加fallback菜单以防数据为空'
        ]
      });
    }
    
    if (permissionFilteredMenus.length < frontendMenus.length) {
      issues.push('🔒 部分菜单被权限过滤');
      solutions.push({
        problem: '权限过滤过严',
        solution: '调整权限检查逻辑',
        implementation: [
          '1. 检查用户角色和权限配置',
          '2. 调整菜单的权限要求',
          '3. 为管理员用户提供更宽松的权限检查',
          '4. 添加权限调试信息'
        ]
      });
    }
    
    console.log('⚠️ 发现的问题:');
    issues.forEach((issue, index) => {
      console.log(`   ${index + 1}. ${issue}`);
    });
    
    console.log('\n💡 解决方案:');
    solutions.forEach((solution, index) => {
      console.log(`\n${index + 1}. ${solution.problem}:`);
      console.log(`   解决方案: ${solution.solution}`);
      console.log('   实施步骤:');
      solution.implementation.forEach((step, stepIndex) => {
        console.log(`     ${step}`);
      });
    });
    
    // 7. 生成修复代码建议
    console.log('\n7️⃣ 修复代码建议...');
    
    console.log('📝 建议的修复步骤:');
    console.log('\n1. 创建MenuContext (frontend/web-app/src/contexts/MenuContext.tsx):');
    console.log(`
import React, { createContext, useContext, useState, useEffect } from 'react';
import { frontendMenuService } from '../services/frontendMenuService';

const MenuContext = createContext(null);

export const MenuProvider = ({ children }) => {
  const [menuItems, setMenuItems] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const loadMenuData = async (userPermissions) => {
    try {
      setLoading(true);
      const menus = await frontendMenuService.getMainMenu(userPermissions);
      setMenuItems(menus);
      setError(null);
    } catch (err) {
      setError(err.message);
      setMenuItems([]);
    } finally {
      setLoading(false);
    }
  };

  return (
    <MenuContext.Provider value={{ menuItems, loading, error, loadMenuData }}>
      {children}
    </MenuContext.Provider>
  );
};

export const useMenu = () => {
  const context = useContext(MenuContext);
  if (!context) {
    throw new Error('useMenu must be used within a MenuProvider');
  }
  return context;
};
    `);
    
    console.log('\n2. 修改Sidebar组件，使用MenuContext而不是直接调用API');
    console.log('\n3. 修改DynamicMenu组件，使用MenuContext而不是直接调用API');
    console.log('\n4. 在App.tsx中添加MenuProvider包装器');
    
    console.log('\n🎉 菜单问题分析完成！');
    
  } catch (error) {
    console.error('❌ 修复过程中发生错误:', error.message);
  }
}

// 运行修复分析
fixMenuDisplayIssue();