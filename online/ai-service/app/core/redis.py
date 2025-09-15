#!/usr/bin/env python3
"""
码道 AI智能体服务Redis模块

管理Redis连接，提供缓存、会话存储等功能。
"""

import json
import asyncio
from typing import Any, Optional, Union
from redis.asyncio import Redis, ConnectionPool
from loguru import logger

from .config import settings


class RedisManager:
    """Redis连接管理器"""
    
    def __init__(self):
        self.redis: Optional[Redis] = None
        self.pool: Optional[ConnectionPool] = None
    
    async def init_redis(self) -> None:
        """初始化Redis连接"""
        try:
            # 创建连接池
            self.pool = ConnectionPool.from_url(
                settings.REDIS_URL,
                password=settings.REDIS_PASSWORD,
                db=settings.REDIS_DB,
                max_connections=20,
                retry_on_timeout=True,
                decode_responses=True
            )
            
            # 创建Redis客户端
            self.redis = Redis(connection_pool=self.pool)
            
            # 测试连接
            await self.redis.ping()
            logger.info("✓ Redis连接已建立")
            
        except Exception as e:
            logger.error(f"Redis连接失败: {e}")
            raise
    
    async def close_redis(self) -> None:
        """关闭Redis连接"""
        if self.redis:
            await self.redis.close()
            logger.info("✓ Redis连接已关闭")
    
    async def get(self, key: str) -> Optional[str]:
        """获取缓存值"""
        try:
            return await self.redis.get(key)
        except Exception as e:
            logger.error(f"Redis GET错误 {key}: {e}")
            return None
    
    async def set(
        self, 
        key: str, 
        value: Union[str, dict, list], 
        expire: Optional[int] = None
    ) -> bool:
        """设置缓存值"""
        try:
            # 如果值是字典或列表，序列化为JSON
            if isinstance(value, (dict, list)):
                value = json.dumps(value, ensure_ascii=False)
            
            result = await self.redis.set(key, value, ex=expire)
            return bool(result)
        except Exception as e:
            logger.error(f"Redis SET错误 {key}: {e}")
            return False
    
    async def delete(self, key: str) -> bool:
        """删除缓存键"""
        try:
            result = await self.redis.delete(key)
            return bool(result)
        except Exception as e:
            logger.error(f"Redis DELETE错误 {key}: {e}")
            return False
    
    async def exists(self, key: str) -> bool:
        """检查键是否存在"""
        try:
            result = await self.redis.exists(key)
            return bool(result)
        except Exception as e:
            logger.error(f"Redis EXISTS错误 {key}: {e}")
            return False
    
    async def expire(self, key: str, seconds: int) -> bool:
        """设置键的过期时间"""
        try:
            result = await self.redis.expire(key, seconds)
            return bool(result)
        except Exception as e:
            logger.error(f"Redis EXPIRE错误 {key}: {e}")
            return False
    
    async def ttl(self, key: str) -> int:
        """获取键的剩余生存时间"""
        try:
            return await self.redis.ttl(key)
        except Exception as e:
            logger.error(f"Redis TTL错误 {key}: {e}")
            return -1
    
    async def incr(self, key: str, amount: int = 1) -> Optional[int]:
        """递增计数器"""
        try:
            return await self.redis.incr(key, amount)
        except Exception as e:
            logger.error(f"Redis INCR错误 {key}: {e}")
            return None
    
    async def get_json(self, key: str) -> Optional[Union[dict, list]]:
        """获取JSON格式的缓存值"""
        try:
            value = await self.get(key)
            if value:
                return json.loads(value)
            return None
        except json.JSONDecodeError as e:
            logger.error(f"JSON解析错误 {key}: {e}")
            return None
    
    async def set_json(
        self, 
        key: str, 
        value: Union[dict, list], 
        expire: Optional[int] = None
    ) -> bool:
        """设置JSON格式的缓存值"""
        return await self.set(key, value, expire)
    
    async def health_check(self) -> bool:
        """Redis健康检查"""
        try:
            await self.redis.ping()
            return True
        except Exception as e:
            logger.error(f"Redis健康检查失败: {e}")
            return False


# 全局Redis管理器实例
redis_manager = RedisManager()


# 便捷函数
async def init_redis() -> None:
    """初始化Redis连接"""
    await redis_manager.init_redis()


async def close_redis() -> None:
    """关闭Redis连接"""
    await redis_manager.close_redis()


async def get_redis() -> Redis:
    """获取Redis客户端实例"""
    if not redis_manager.redis:
        await redis_manager.init_redis()
    return redis_manager.redis


# 缓存装饰器
def cache_result(key_prefix: str, expire: int = 3600):
    """缓存函数结果的装饰器"""
    def decorator(func):
        async def wrapper(*args, **kwargs):
            # 生成缓存键
            cache_key = f"{key_prefix}:{hash(str(args) + str(kwargs))}"
            
            # 尝试从缓存获取
            cached_result = await redis_manager.get_json(cache_key)
            if cached_result is not None:
                logger.debug(f"缓存命中: {cache_key}")
                return cached_result
            
            # 执行函数并缓存结果
            result = await func(*args, **kwargs)
            await redis_manager.set_json(cache_key, result, expire)
            logger.debug(f"缓存设置: {cache_key}")
            
            return result
        return wrapper
    return decorator


if __name__ == "__main__":
    # 测试Redis连接
    async def test_redis():
        logger.info("测试Redis连接...")
        await init_redis()
        
        # 测试基本操作
        await redis_manager.set("test_key", "test_value", 60)
        value = await redis_manager.get("test_key")
        logger.info(f"测试值: {value}")
        
        # 测试JSON操作
        test_data = {"message": "Hello Redis", "timestamp": "2024-01-15"}
        await redis_manager.set_json("test_json", test_data, 60)
        json_value = await redis_manager.get_json("test_json")
        logger.info(f"JSON值: {json_value}")
        
        # 健康检查
        health = await redis_manager.health_check()
        if health:
            logger.info("✓ Redis连接正常")
        else:
            logger.error("✗ Redis连接失败")
        
        await close_redis()
    
    asyncio.run(test_redis())