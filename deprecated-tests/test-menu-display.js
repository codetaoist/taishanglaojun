// 测试菜单显示功能
console.log('🧪 开始测试菜单显示功能...');

// 等待页面加载完成
function waitForPageLoad() {
  return new Promise((resolve) => {
    if (document.readyState === 'complete') {
      resolve();
    } else {
      window.addEventListener('load', resolve);
    }
  });
}

// 等待React应用加载
function waitForReactApp() {
  return new Promise((resolve) => {
    const checkReact = () => {
      if (window.React || document.querySelector('[data-reactroot]') || document.querySelector('#root > div')) {
        resolve();
      } else {
        setTimeout(checkReact, 100);
      }
    };
    checkReact();
  });
}

// 测试菜单元素
async function testMenuElements() {
  console.log('🔍 检查菜单元素...');
  
  // 检查侧边栏
  const sidebar = document.querySelector('.ant-layout-sider, [class*="sidebar"], [class*="Sidebar"]');
  console.log('📋 侧边栏元素:', sidebar);
  
  // 检查菜单容器
  const menuContainer = document.querySelector('.ant-menu, [class*="menu"], [class*="Menu"]');
  console.log('📋 菜单容器:', menuContainer);
  
  // 检查菜单项
  const menuItems = document.querySelectorAll('.ant-menu-item, [class*="menu-item"], [class*="MenuItem"]');
  console.log('📋 菜单项数量:', menuItems.length);
  
  if (menuItems.length > 0) {
    console.log('📋 菜单项列表:');
    menuItems.forEach((item, index) => {
      console.log(`  ${index + 1}. ${item.textContent?.trim() || '(无文本)'}`);
    });
  } else {
    console.log('❌ 没有找到菜单项');
  }
  
  // 检查加载状态
  const loadingElements = document.querySelectorAll('[class*="loading"], [class*="Loading"], .ant-spin');
  console.log('⏳ 加载元素数量:', loadingElements.length);
  
  // 检查错误信息
  const errorElements = document.querySelectorAll('[class*="error"], [class*="Error"]');
  console.log('❌ 错误元素数量:', errorElements.length);
  
  return {
    hasSidebar: !!sidebar,
    hasMenuContainer: !!menuContainer,
    menuItemsCount: menuItems.length,
    isLoading: loadingElements.length > 0,
    hasErrors: errorElements.length > 0
  };
}

// 检查控制台日志
function checkConsoleLogs() {
  console.log('📝 检查控制台日志...');
  
  // 保存原始的console方法
  const originalLog = console.log;
  const originalError = console.error;
  const originalWarn = console.warn;
  
  const logs = [];
  
  // 拦截console输出
  console.log = (...args) => {
    logs.push({ type: 'log', args });
    originalLog.apply(console, args);
  };
  
  console.error = (...args) => {
    logs.push({ type: 'error', args });
    originalError.apply(console, args);
  };
  
  console.warn = (...args) => {
    logs.push({ type: 'warn', args });
    originalWarn.apply(console, args);
  };
  
  // 等待一段时间收集日志
  setTimeout(() => {
    console.log = originalLog;
    console.error = originalError;
    console.warn = originalWarn;
    
    console.log('📊 控制台日志统计:');
    console.log(`  普通日志: ${logs.filter(l => l.type === 'log').length}`);
    console.log(`  错误日志: ${logs.filter(l => l.type === 'error').length}`);
    console.log(`  警告日志: ${logs.filter(l => l.type === 'warn').length}`);
    
    // 查找菜单相关的日志
    const menuLogs = logs.filter(l => 
      l.args.some(arg => 
        typeof arg === 'string' && 
        (arg.includes('菜单') || arg.includes('menu') || arg.includes('Menu'))
      )
    );
    
    if (menuLogs.length > 0) {
      console.log('🔍 菜单相关日志:');
      menuLogs.forEach((log, index) => {
        console.log(`  ${index + 1}. [${log.type}]`, ...log.args);
      });
    }
  }, 3000);
}

// 主测试函数
async function runTest() {
  try {
    console.log('⏳ 等待页面加载...');
    await waitForPageLoad();
    
    console.log('⏳ 等待React应用加载...');
    await waitForReactApp();
    
    console.log('✅ 页面加载完成，开始测试...');
    
    // 开始监控控制台日志
    checkConsoleLogs();
    
    // 等待一段时间让应用初始化
    await new Promise(resolve => setTimeout(resolve, 2000));
    
    // 测试菜单元素
    const testResult = await testMenuElements();
    
    console.log('📊 测试结果:', testResult);
    
    if (testResult.menuItemsCount > 0) {
      console.log('✅ 菜单显示正常！');
    } else if (testResult.isLoading) {
      console.log('⏳ 菜单正在加载中...');
      // 等待加载完成后再次测试
      setTimeout(async () => {
        const retestResult = await testMenuElements();
        console.log('📊 重新测试结果:', retestResult);
      }, 3000);
    } else {
      console.log('❌ 菜单显示异常！');
    }
    
  } catch (error) {
    console.error('❌ 测试过程中出错:', error);
  }
}

// 启动测试
runTest();