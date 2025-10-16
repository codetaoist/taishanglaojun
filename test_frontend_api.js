// 测试前端API调用
// 在浏览器控制台中运行此脚本

// 首先设置token
const token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiZmI2NzNmOTAtZTc1My00NGU4LWE1ZWItNzUxNmQ0YzY4M2FkIiwidXNlcm5hbWUiOiJhZG1pbiIsInJvbGUiOiJBRE1JTiIsImxldmVsIjo1LCJpc3MiOiJ0YWlzaGFuZy1sYW9qdW4iLCJleHAiOjE3NjA2Nzc4NDAsIm5iZiI6MTc2MDU5MTQ0MCwiaWF0IjoxNzYwNTkxNDQwfQ.zyAlXixegNqdtyJ0kF279CBRgvkXfLbSBwMMgiWLjMg";

localStorage.setItem('auth_token', token);
localStorage.setItem('token', token);

const userInfo = {
  user_id: "fb673f90-e753-44e8-a5eb-7516d4c683ad",
  username: "admin",
  email: "admin@example.com",
  role: "ADMIN"
};

localStorage.setItem('user', JSON.stringify(userInfo));

console.log('✅ Token已设置');

// 测试API调用
async function testRoleAPI() {
  try {
    console.log('🔄 测试角色列表API...');
    
    const response = await fetch('http://localhost:8080/api/v1/roles', {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    });
    
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }
    
    const data = await response.json();
    console.log('✅ API响应成功:', data);
    console.log(`📊 角色总数: ${data.total}`);
    console.log(`📋 角色列表长度: ${data.data?.length || 0}`);
    
    // 检查角色名称
    if (data.data && Array.isArray(data.data)) {
      console.log('🏷️ 角色名称列表:');
      data.data.forEach((role, index) => {
        console.log(`  ${index + 1}. ${role.name} (${role.code}) - ${role.type}`);
      });
    }
    
    return data;
  } catch (error) {
    console.error('❌ API调用失败:', error);
    return null;
  }
}

// 测试查询功能
async function testRoleQuery() {
  try {
    console.log('🔍 测试角色查询功能...');
    
    // 测试搜索
    const searchResponse = await fetch('http://localhost:8080/api/v1/roles?search=admin', {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    });
    
    if (searchResponse.ok) {
      const searchData = await searchResponse.json();
      console.log('✅ 搜索功能正常:', searchData);
    }
    
    // 测试类型筛选
    const typeResponse = await fetch('http://localhost:8080/api/v1/roles?type=system', {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    });
    
    if (typeResponse.ok) {
      const typeData = await typeResponse.json();
      console.log('✅ 类型筛选功能正常:', typeData);
    }
    
  } catch (error) {
    console.error('❌ 查询功能测试失败:', error);
  }
}

// 运行测试
console.log('🚀 开始测试...');
testRoleAPI().then(() => {
  testRoleQuery().then(() => {
    console.log('✨ 测试完成，请刷新页面查看角色管理页面');
  });
});