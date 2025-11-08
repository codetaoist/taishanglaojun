import uvicorn
from fastapi import FastAPI
from app.core.config import settings
from app.utils.logger import configure_logging, get_logger

# 配置日志
configure_logging()
logger = get_logger(__name__)

# 创建FastAPI应用
app = FastAPI(
    title=settings.app_name,
    version=settings.app_version,
    description="Taishang AI微服务，提供向量数据库和模型调用功能",
    debug=settings.debug
)


@app.get("/health")
async def health_check():
    """健康检查端点"""
    return {"status": "healthy", "service": settings.app_name}


@app.get("/")
async def root():
    """根端点"""
    return {
        "message": f"Welcome to {settings.app_name}",
        "version": settings.app_version
    }


if __name__ == "__main__":
    logger.info("Starting Taishang AI Service", 
                http_port=settings.http_port, 
                grpc_port=settings.grpc_port)
    
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=settings.http_port,
        reload=settings.debug,
        log_level=settings.log_level
    )