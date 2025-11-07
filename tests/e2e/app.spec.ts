import { test, expect } from '@playwright/test';

// 测试基础设置
const BASE_URL = process.env.BASE_URL || 'http://localhost:5173';

// 测试数据
const testUser = {
  username: 'testuser',
  password: 'testpassword',
  email: 'test@example.com'
};

test.describe('前端应用集成测试', () => {
  test.beforeEach(async ({ page }) => {
    // 每个测试前访问应用首页
    await page.goto(BASE_URL);
  });

  test('应用首页加载正常', async ({ page }) => {
    // 等待页面加载完成
    await page.waitForLoadState('networkidle');
    
    // 检查页面标题
    await expect(page).toHaveTitle(/太上老君/);
    
    // 检查主要导航元素是否存在
    await expect(page.locator('nav')).toBeVisible();
    await expect(page.locator('nav a[href="/"]')).toBeVisible();
    await expect(page.locator('nav a[href="/dashboard"]')).toBeVisible();
    await expect(page.locator('nav a[href="/taishang"]')).toBeVisible();
    await expect(page.locator('nav a[href="/laojun"]')).toBeVisible();
  });

  test('仪表板页面加载正常', async ({ page }) => {
    // 点击仪表板链接
    await page.click('nav a[href="/dashboard"]');
    
    // 等待页面加载
    await page.waitForLoadState('networkidle');
    
    // 检查URL是否正确
    await expect(page).toHaveURL(`${BASE_URL}/dashboard`);
    
    // 检查仪表板内容
    await expect(page.locator('h1')).toContainText('仪表板');
  });

  test('太上域页面加载正常', async ({ page }) => {
    // 点击太上域链接
    await page.click('nav a[href="/taishang"]');
    
    // 等待页面加载
    await page.waitForLoadState('networkidle');
    
    // 检查URL是否正确
    await expect(page).toHaveURL(`${BASE_URL}/taishang`);
    
    // 检查太上域内容
    await expect(page.locator('h1')).toContainText('太上域');
  });

  test('老君域页面加载正常', async ({ page }) => {
    // 点击老君域链接
    await page.click('nav a[href="/laojun"]');
    
    // 等待页面加载
    await page.waitForLoadState('networkidle');
    
    // 检查URL是否正确
    await expect(page).toHaveURL(`${BASE_URL}/laojun`);
    
    // 检查老君域内容
    await expect(page.locator('h1')).toContainText('老君域');
  });

  test('API错误处理', async ({ page }) => {
    // 监听网络请求
    page.route('**/api/**', route => {
      // 模拟API错误响应
      route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({
          code: 500,
          message: '服务器内部错误'
        })
      });
    });

    // 导航到会触发API请求的页面
    await page.click('nav a[href="/dashboard"]');
    
    // 等待错误提示显示
    await expect(page.locator('.error-message')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('.error-message')).toContainText('服务器内部错误');
  });

  test('响应式设计测试', async ({ page }) => {
    // 测试桌面视图
    await page.setViewportSize({ width: 1200, height: 800 });
    await expect(page.locator('nav')).toBeVisible();
    
    // 测试平板视图
    await page.setViewportSize({ width: 768, height: 1024 });
    await expect(page.locator('nav')).toBeVisible();
    
    // 测试手机视图
    await page.setViewportSize({ width: 375, height: 667 });
    await expect(page.locator('nav')).toBeVisible();
    
    // 在手机视图中，检查是否有汉堡菜单
    await expect(page.locator('.mobile-menu-toggle')).toBeVisible();
    
    // 点击汉堡菜单，检查导航项是否显示
    await page.click('.mobile-menu-toggle');
    await expect(page.locator('.mobile-menu')).toBeVisible();
  });

  test('表单交互测试', async ({ page }) => {
    // 导航到有表单的页面
    await page.click('nav a[href="/taishang"]');
    
    // 等待页面加载
    await page.waitForLoadState('networkidle');
    
    // 查找并点击添加按钮
    await page.click('button:has-text("添加")');
    
    // 等待模态框或表单显示
    await expect(page.locator('.modal, .form')).toBeVisible();
    
    // 填写表单
    await page.fill('input[name="name"]', '测试域名称');
    await page.fill('textarea[name="description"]', '这是一个测试域的描述');
    
    // 提交表单
    await page.click('button:has-text("提交")');
    
    // 检查是否有成功提示
    await expect(page.locator('.success-message')).toBeVisible({ timeout: 5000 });
  });
});

test.describe('API集成测试', () => {
  test('API健康检查', async ({ request }) => {
    const response = await request.get(`${BASE_URL.replace('5173', '8080')}/health`);
    expect(response.status()).toBe(200);
    
    const data = await response.json();
    expect(data.status).toBe('ok');
  });

  test('太上域API测试', async ({ request }) => {
    // 获取太上域列表
    const domainsResponse = await request.get(`${BASE_URL.replace('5173', '8080')}/api/taishang/domains`);
    expect([200, 401]).toContain(domainsResponse.status()); // 可能成功或需要认证
    
    if (domainsResponse.status() === 200) {
      const domainsData = await domainsResponse.json();
      expect(Array.isArray(domainsData.data)).toBe(true);
    }
  });

  test('老君域API测试', async ({ request }) => {
    // 获取老君域列表
    const domainsResponse = await request.get(`${BASE_URL.replace('5173', '8080')}/api/laojun/domains`);
    expect([200, 401]).toContain(domainsResponse.status()); // 可能成功或需要认证
    
    if (domainsResponse.status() === 200) {
      const domainsData = await domainsResponse.json();
      expect(Array.isArray(domainsData.data)).toBe(true);
    }
  });
});