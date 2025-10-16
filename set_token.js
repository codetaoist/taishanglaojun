// 在浏览器控制台中运行此脚本来设置token
const token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiZmI2NzNmOTAtZTc1My00NGU4LWE1ZWItNzUxNmQ0YzY4M2FkIiwidXNlcm5hbWUiOiJhZG1pbiIsInJvbGUiOiJBRE1JTiIsImxldmVsIjo1LCJpc3MiOiJ0YWlzaGFuZy1sYW9qdW4iLCJleHAiOjE3NjA2Nzc4NDAsIm5iZiI6MTc2MDU5MTQ0MCwiaWF0IjoxNzYwNTkxNDQwfQ.zyAlXixegNqdtyJ0kF279CBRgvkXfLbSBwMMgiWLjMg";

// 设置token到localStorage
localStorage.setItem('auth_token', token);
localStorage.setItem('token', token);

// 设置用户信息
const userInfo = {
  user_id: "fb673f90-e753-44e8-a5eb-7516d4c683ad",
  username: "admin",
  email: "admin@example.com",
  role: "ADMIN"
};

localStorage.setItem('user', JSON.stringify(userInfo));

console.log('Token已设置:', token);
console.log('用户信息已设置:', userInfo);
console.log('请刷新页面以应用更改');