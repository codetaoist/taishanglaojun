#!/usr/bin/env python3
"""
码道 (Code Taoist) AI智能体服务

基于FastAPI和LangChain构建的智能编程助手服务，
提供自然语言编程指令处理、代码生成、知识库查询等功能。
"""

import os
import sys
from contextlib import asynccontextmanager
from typing import Dict, Any

from fastapi import FastAPI, HTTPException, Depends, status
from fastapi.middleware.cors import CORSMiddleware
from fastapi.middleware.trustedhost import TrustedHostMiddleware
from fastapi.responses import JSONResponse
from loguru import logger
from prometheus_client import make_asgi_app

from app.core.config import settings
from app.core.database import init_db
from app.core.redis import init_redis
from app.api.v1 import api_router
from app.core.auth import get_current_user
from app.core.monitoring import setup_monitoring


@asynccontextmanager
async def lifespan(app: FastAPI):
    """应用生命周期管理"""
    # 启动时初始化
    logger.info("🚀 启动码道AI智能体服务...")
    
    # 初始化数据库
    await init_db()
    logger.info("✓ 数据库连接已建立")
    
    # 初始化Redis
    await init_redis()
    logger.info("✓ Redis连接已建立")
    
    # 设置监控
    setup_monitoring()
    logger.info("✓ 监控系统已启动")
    
    logger.info("🎉 码道AI智能体服务启动完成")
    
    yield
    
    # 关闭时清理
    logger.info("🔄 正在关闭码道AI智能体服务...")
    logger.info("✓ 服务已安全关闭")


# 创建FastAPI应用
app = FastAPI(
    title="码道 AI智能体服务",
    description="基于LangChain的智能编程助手API服务",
    version="1.0.0",
    docs_url="/docs" if settings.DEBUG else None,
    redoc_url="/redoc" if settings.DEBUG else None,
    lifespan=lifespan
)

# 添加CORS中间件
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.ALLOWED_HOSTS,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# 添加可信主机中间件
app.add_middleware(
    TrustedHostMiddleware,
    allowed_hosts=settings.ALLOWED_HOSTS
)

# 添加API路由
app.include_router(api_router, prefix="/api/v1")

# 添加Prometheus监控端点
metrics_app = make_asgi_app()
app.mount("/metrics", metrics_app)


@app.get("/health")
async def health_check() -> Dict[str, Any]:
    """健康检查端点"""
    return {
        "status": "healthy",
        "service": "ai-service",
        "version": "1.0.0",
        "timestamp": "2024-01-15T12:00:00Z"
    }


@app.get("/")
async def root() -> Dict[str, str]:
    """根路径"""
    return {
        "message": "码道 (Code Taoist) AI智能体服务",
        "docs": "/docs",
        "health": "/health"
    }


@app.exception_handler(HTTPException)
async def http_exception_handler(request, exc: HTTPException):
    """HTTP异常处理器"""
    logger.error(f"HTTP异常: {exc.status_code} - {exc.detail}")
    return JSONResponse(
        status_code=exc.status_code,
        content={
            "error": exc.detail,
            "status_code": exc.status_code,
            "timestamp": "2024-01-15T12:00:00Z"
        }
    )


@app.exception_handler(Exception)
async def general_exception_handler(request, exc: Exception):
    """通用异常处理器"""
    logger.error(f"未处理的异常: {type(exc).__name__} - {str(exc)}")
    return JSONResponse(
        status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
        content={
            "error": "内部服务器错误",
            "status_code": 500,
            "timestamp": "2024-01-15T12:00:00Z"
        }
    )


if __name__ == "__main__":
    import uvicorn
    
    # 配置日志
    logger.remove()
    logger.add(
        sys.stdout,
        format="<green>{time:YYYY-MM-DD HH:mm:ss}</green> | <level>{level: <8}</level> | <cyan>{name}</cyan>:<cyan>{function}</cyan>:<cyan>{line}</cyan> - <level>{message}</level>",
        level="INFO" if not settings.DEBUG else "DEBUG"
    )
    
    # 启动服务
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=8000,
        reload=settings.DEBUG,
        log_level="info"
    )