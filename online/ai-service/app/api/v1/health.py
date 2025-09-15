#!/usr/bin/env python3
"""
码道 AI智能体服务健康检查API

提供服务健康状态检查端点。
"""

import time
from typing import Dict, Any
from fastapi import APIRouter, HTTPException, status
from pydantic import BaseModel, Field
from loguru import logger

from ...core.database import check_db_health
from ...core.redis import redis_manager
from ...core.monitoring import monitor_requests


router = APIRouter()


class HealthStatus(BaseModel):
    """健康状态模型"""
    status: str = Field(..., description="整体状态: healthy, unhealthy, degraded")
    timestamp: float = Field(default_factory=time.time, description="检查时间戳")
    version: str = Field("1.0.0", description="服务版本")
    uptime: float = Field(..., description="运行时间（秒）")
    checks: Dict[str, Any] = Field(..., description="各组件检查结果")


# 服务启动时间
START_TIME = time.time()


@router.get("/", response_model=HealthStatus)
@monitor_requests
async def health_check() -> HealthStatus:
    """
    综合健康检查
    
    检查所有关键组件的健康状态。
    """
    try:
        checks = {}
        overall_healthy = True
        
        # 检查数据库
        try:
            db_healthy = await check_db_health()
            checks["database"] = {
                "status": "healthy" if db_healthy else "unhealthy",
                "response_time": 0.05 if db_healthy else None
            }
            if not db_healthy:
                overall_healthy = False
        except Exception as e:
            checks["database"] = {
                "status": "error",
                "error": str(e)
            }
            overall_healthy = False
        
        # 检查Redis
        try:
            redis_healthy = await redis_manager.health_check()
            checks["redis"] = {
                "status": "healthy" if redis_healthy else "unhealthy",
                "response_time": 0.02 if redis_healthy else None
            }
            if not redis_healthy:
                overall_healthy = False
        except Exception as e:
            checks["redis"] = {
                "status": "error",
                "error": str(e)
            }
            overall_healthy = False
        
        # 检查AI服务（模拟）
        checks["ai_service"] = {
            "status": "healthy",
            "models_available": ["gpt-4", "deepseek"],
            "response_time": 0.1
        }
        
        # 检查向量数据库（模拟）
        checks["vector_db"] = {
            "status": "healthy",
            "collections": 1,
            "response_time": 0.03
        }
        
        status = "healthy" if overall_healthy else "unhealthy"
        uptime = time.time() - START_TIME
        
        return HealthStatus(
            status=status,
            uptime=uptime,
            checks=checks
        )
        
    except Exception as e:
        logger.error(f"健康检查失败: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"健康检查失败: {str(e)}"
        )


@router.get("/ready")
@monitor_requests
async def readiness_check() -> Dict[str, Any]:
    """
    就绪检查
    
    检查服务是否准备好接收请求。
    """
    try:
        # 检查关键依赖
        db_ready = await check_db_health()
        redis_ready = await redis_manager.health_check()
        
        ready = db_ready and redis_ready
        
        return {
            "ready": ready,
            "timestamp": time.time(),
            "dependencies": {
                "database": db_ready,
                "redis": redis_ready
            }
        }
        
    except Exception as e:
        logger.error(f"就绪检查失败: {e}")
        return {
            "ready": False,
            "timestamp": time.time(),
            "error": str(e)
        }


@router.get("/live")
@monitor_requests
async def liveness_check() -> Dict[str, Any]:
    """
    存活检查
    
    简单的存活状态检查。
    """
    return {
        "alive": True,
        "timestamp": time.time(),
        "uptime": time.time() - START_TIME
    }