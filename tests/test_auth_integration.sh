#!/bin/bash

# 测试认证服务和API服务的集成脚本

echo "开始测试认证服务和API服务的集成..."

# 测试认证服务健康检查
echo "1. 测试认证服务健康检查..."
curl -s http://localhost:8081/health | jq .
echo ""

# 测试用户注册
echo "2. 测试用户注册..."
REGISTER_RESPONSE=$(curl -s -X POST http://localhost:8081/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }')

echo "$REGISTER_RESPONSE" | jq .
echo ""

# 提取token
ACCESS_TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.data.access_token')
REFRESH_TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.data.refresh_token')

echo "Access Token: $ACCESS_TOKEN"
echo "Refresh Token: $REFRESH_TOKEN"
echo ""

# 测试用户登录
echo "3. 测试用户登录..."
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }')

echo "$LOGIN_RESPONSE" | jq .
echo ""

# 更新token
ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.access_token')
REFRESH_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.refresh_token')

# 测试获取用户信息
echo "4. 测试获取用户信息..."
PROFILE_RESPONSE=$(curl -s -X GET http://localhost:8081/api/v1/profile \
  -H "Authorization: Bearer $ACCESS_TOKEN")

echo "$PROFILE_RESPONSE" | jq .
echo ""

# 测试API服务的受保护端点
echo "5. 测试API服务的受保护端点..."
API_RESPONSE=$(curl -s -X GET http://localhost:8082/api/v1/protected \
  -H "Authorization: Bearer $ACCESS_TOKEN")

echo "$API_RESPONSE" | jq .
echo ""

# 测试令牌刷新
echo "6. 测试令牌刷新..."
REFRESH_RESPONSE=$(curl -s -X POST http://localhost:8081/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{
    \"refresh_token\": \"$REFRESH_TOKEN\"
  }")

echo "$REFRESH_RESPONSE" | jq .
echo ""

# 测试用户登出
echo "7. 测试用户登出..."
LOGOUT_RESPONSE=$(curl -s -X POST http://localhost:8081/api/v1/auth/logout \
  -H "Authorization: Bearer $ACCESS_TOKEN")

echo "$LOGOUT_RESPONSE" | jq .
echo ""

# 测试管理员登录
echo "8. 测试管理员登录..."
ADMIN_LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@taishanglaojun.com",
    "password": "admin123"
  }')

echo "$ADMIN_LOGIN_RESPONSE" | jq .
echo ""

# 更新管理员token
ADMIN_ACCESS_TOKEN=$(echo "$ADMIN_LOGIN_RESPONSE" | jq -r '.data.access_token')

# 测试管理员端点
echo "9. 测试管理员端点..."
ADMIN_RESPONSE=$(curl -s -X GET http://localhost:8081/api/v1/admin/users/1 \
  -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN")

echo "$ADMIN_RESPONSE" | jq .
echo ""

echo "测试完成！"