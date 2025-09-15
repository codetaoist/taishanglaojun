#!/usr/bin/env python3
"""
码道 AI智能体服务认证模块

基于JWT和Keycloak的用户认证和授权。
"""

import jwt
from datetime import datetime, timedelta
from typing import Optional, Dict, Any
from fastapi import HTTPException, Depends, status
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from loguru import logger

from .config import settings
from .redis import redis_manager


security = HTTPBearer()


class AuthManager:
    """认证管理器"""
    
    def __init__(self):
        self.secret_key = settings.SECRET_KEY
        self.algorithm = settings.JWT_ALGORITHM
        self.expire_minutes = settings.JWT_EXPIRE_MINUTES
    
    def create_access_token(self, data: Dict[str, Any]) -> str:
        """创建访问令牌"""
        to_encode = data.copy()
        expire = datetime.utcnow() + timedelta(minutes=self.expire_minutes)
        to_encode.update({"exp": expire})
        
        encoded_jwt = jwt.encode(
            to_encode, 
            self.secret_key, 
            algorithm=self.algorithm
        )
        return encoded_jwt
    
    def verify_token(self, token: str) -> Optional[Dict[str, Any]]:
        """验证访问令牌"""
        try:
            payload = jwt.decode(
                token, 
                self.secret_key, 
                algorithms=[self.algorithm]
            )
            return payload
        except jwt.ExpiredSignatureError:
            logger.warning("JWT令牌已过期")
            return None
        except jwt.JWTError as e:
            logger.warning(f"JWT验证失败: {e}")
            return None
    
    async def get_user_from_token(self, token: str) -> Optional[Dict[str, Any]]:
        """从令牌获取用户信息"""
        payload = self.verify_token(token)
        if not payload:
            return None
        
        user_id = payload.get("sub")
        if not user_id:
            return None
        
        # 从缓存获取用户信息
        user_cache_key = f"user:{user_id}"
        user_info = await redis_manager.get_json(user_cache_key)
        
        if user_info:
            return user_info
        
        # 如果缓存中没有，返回基本信息
        return {
            "user_id": user_id,
            "username": payload.get("preferred_username"),
            "email": payload.get("email"),
            "roles": payload.get("realm_access", {}).get("roles", []),
            "projects": payload.get("projects", [])
        }
    
    async def cache_user_info(self, user_id: str, user_info: Dict[str, Any]) -> None:
        """缓存用户信息"""
        cache_key = f"user:{user_id}"
        await redis_manager.set_json(cache_key, user_info, expire=3600)  # 1小时
    
    async def invalidate_user_cache(self, user_id: str) -> None:
        """清除用户缓存"""
        cache_key = f"user:{user_id}"
        await redis_manager.delete(cache_key)
    
    def check_permissions(self, user_roles: list, required_roles: list) -> bool:
        """检查用户权限"""
        return any(role in user_roles for role in required_roles)


# 全局认证管理器实例
auth_manager = AuthManager()


async def get_current_user(
    credentials: HTTPAuthorizationCredentials = Depends(security)
) -> Dict[str, Any]:
    """获取当前用户依赖"""
    token = credentials.credentials
    
    user = await auth_manager.get_user_from_token(token)
    if not user:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="无效的认证令牌",
            headers={"WWW-Authenticate": "Bearer"},
        )
    
    return user


async def get_current_active_user(
    current_user: Dict[str, Any] = Depends(get_current_user)
) -> Dict[str, Any]:
    """获取当前活跃用户依赖"""
    if current_user.get("disabled"):
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="用户账号已被禁用"
        )
    return current_user


def require_roles(required_roles: list):
    """要求特定角色的装饰器"""
    def role_checker(current_user: Dict[str, Any] = Depends(get_current_active_user)):
        user_roles = current_user.get("roles", [])
        
        if not auth_manager.check_permissions(user_roles, required_roles):
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail=f"需要以下角色之一: {', '.join(required_roles)}"
            )
        
        return current_user
    
    return role_checker


def require_project_access(project_id: str):
    """要求项目访问权限的装饰器"""
    def project_checker(current_user: Dict[str, Any] = Depends(get_current_active_user)):
        user_projects = current_user.get("projects", [])
        user_roles = current_user.get("roles", [])
        
        # 管理员可以访问所有项目
        if "admin" in user_roles:
            return current_user
        
        # 检查用户是否有该项目的访问权限
        if project_id not in user_projects:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="没有访问该项目的权限"
            )
        
        return current_user
    
    return project_checker


class RateLimiter:
    """API限流器"""
    
    def __init__(self, requests: int = 100, window: int = 60):
        self.requests = requests
        self.window = window
    
    async def check_rate_limit(self, user_id: str) -> bool:
        """检查用户是否超过限流"""
        key = f"rate_limit:{user_id}"
        
        # 获取当前请求计数
        current_requests = await redis_manager.get(key)
        
        if current_requests is None:
            # 第一次请求，设置计数器
            await redis_manager.set(key, "1", expire=self.window)
            return True
        
        current_count = int(current_requests)
        if current_count >= self.requests:
            return False
        
        # 增加计数
        await redis_manager.incr(key)
        return True


# 默认限流器
default_rate_limiter = RateLimiter(
    requests=settings.RATE_LIMIT_REQUESTS,
    window=settings.RATE_LIMIT_WINDOW
)


async def check_rate_limit(
    current_user: Dict[str, Any] = Depends(get_current_active_user)
) -> Dict[str, Any]:
    """检查API限流依赖"""
    user_id = current_user["user_id"]
    
    if not await default_rate_limiter.check_rate_limit(user_id):
        raise HTTPException(
            status_code=status.HTTP_429_TOO_MANY_REQUESTS,
            detail="请求过于频繁，请稍后再试"
        )
    
    return current_user


# 模拟用户数据（实际应该从数据库获取）
MOCK_USERS = {
    "user_123": {
        "user_id": "user_123",
        "username": "demo@codetaoist.com",
        "email": "demo@codetaoist.com",
        "roles": ["developer"],
        "projects": ["proj_1234567890", "proj_1234567891"],
        "disabled": False
    },
    "admin_456": {
        "user_id": "admin_456",
        "username": "admin@codetaoist.com",
        "email": "admin@codetaoist.com",
        "roles": ["admin", "developer"],
        "projects": [],  # 管理员可访问所有项目
        "disabled": False
    }
}


async def create_mock_token(user_id: str) -> str:
    """创建模拟令牌（仅用于开发测试）"""
    if user_id not in MOCK_USERS:
        raise ValueError("用户不存在")
    
    user_info = MOCK_USERS[user_id]
    
    token_data = {
        "sub": user_id,
        "preferred_username": user_info["username"],
        "email": user_info["email"],
        "realm_access": {"roles": user_info["roles"]},
        "projects": user_info["projects"]
    }
    
    # 缓存用户信息
    await auth_manager.cache_user_info(user_id, user_info)
    
    return auth_manager.create_access_token(token_data)


if __name__ == "__main__":
    # 测试认证功能
    import asyncio
    
    async def test_auth():
        logger.info("测试认证功能...")
        
        # 创建测试令牌
        token = await create_mock_token("user_123")
        logger.info(f"生成的令牌: {token[:50]}...")
        
        # 验证令牌
        user = await auth_manager.get_user_from_token(token)
        logger.info(f"用户信息: {user}")
        
        # 测试权限检查
        has_permission = auth_manager.check_permissions(
            user["roles"], ["developer"]
        )
        logger.info(f"开发者权限: {has_permission}")
    
    asyncio.run(test_auth())