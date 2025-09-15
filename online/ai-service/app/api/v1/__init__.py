#!/usr/bin/env python3
"""
码道 AI智能体服务 API v1 路由模块

定义所有API端点的路由配置。
"""

from fastapi import APIRouter

from .chat import router as chat_router
from .knowledge import router as knowledge_router
from .projects import router as projects_router
from .health import router as health_router


# 创建API路由器
api_router = APIRouter()

# 包含各个子路由
api_router.include_router(
    chat_router,
    prefix="/chat",
    tags=["AI聊天"]
)

api_router.include_router(
    knowledge_router,
    prefix="/knowledge",
    tags=["知识库"]
)

api_router.include_router(
    projects_router,
    prefix="/projects",
    tags=["项目管理"]
)

api_router.include_router(
    health_router,
    prefix="/health",
    tags=["健康检查"]
)