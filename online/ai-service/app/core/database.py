#!/usr/bin/env python3
"""
码道 AI智能体服务数据库模块

使用SQLAlchemy管理PostgreSQL数据库连接和会话。
"""

import asyncio
from typing import AsyncGenerator
from sqlalchemy.ext.asyncio import AsyncSession, create_async_engine, async_sessionmaker
from sqlalchemy.orm import DeclarativeBase
from sqlalchemy import MetaData
from loguru import logger

from .config import settings


class Base(DeclarativeBase):
    """SQLAlchemy基础模型类"""
    metadata = MetaData(
        naming_convention={
            "ix": "ix_%(column_0_label)s",
            "uq": "uq_%(table_name)s_%(column_0_name)s",
            "ck": "ck_%(table_name)s_%(constraint_name)s",
            "fk": "fk_%(table_name)s_%(column_0_name)s_%(referred_table_name)s",
            "pk": "pk_%(table_name)s"
        }
    )


# 创建异步数据库引擎
engine = create_async_engine(
    settings.DATABASE_URL.replace("postgresql://", "postgresql+asyncpg://"),
    pool_size=settings.DATABASE_POOL_SIZE,
    max_overflow=settings.DATABASE_MAX_OVERFLOW,
    pool_pre_ping=True,
    echo=settings.DEBUG,
)

# 创建异步会话工厂
AsyncSessionLocal = async_sessionmaker(
    engine,
    class_=AsyncSession,
    expire_on_commit=False
)


async def get_db() -> AsyncGenerator[AsyncSession, None]:
    """获取数据库会话依赖"""
    async with AsyncSessionLocal() as session:
        try:
            yield session
        except Exception as e:
            await session.rollback()
            logger.error(f"数据库会话错误: {e}")
            raise
        finally:
            await session.close()


async def init_db() -> None:
    """初始化数据库"""
    try:
        # 测试数据库连接
        async with engine.begin() as conn:
            # 创建所有表
            await conn.run_sync(Base.metadata.create_all)
            logger.info("✓ 数据库表结构已创建")
            
    except Exception as e:
        logger.error(f"数据库初始化失败: {e}")
        raise


async def close_db() -> None:
    """关闭数据库连接"""
    await engine.dispose()
    logger.info("✓ 数据库连接已关闭")


# 数据库健康检查
async def check_db_health() -> bool:
    """检查数据库连接健康状态"""
    try:
        async with AsyncSessionLocal() as session:
            await session.execute("SELECT 1")
            return True
    except Exception as e:
        logger.error(f"数据库健康检查失败: {e}")
        return False


if __name__ == "__main__":
    # 测试数据库连接
    async def test_connection():
        logger.info("测试数据库连接...")
        await init_db()
        
        health = await check_db_health()
        if health:
            logger.info("✓ 数据库连接正常")
        else:
            logger.error("✗ 数据库连接失败")
            
        await close_db()
    
    asyncio.run(test_connection())